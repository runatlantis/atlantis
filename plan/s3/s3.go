package s3

import (
	"os"
	pathutil "path"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/plan"
	"github.com/pkg/errors"
)

type Backend struct {
	s3         *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
	keyPrefix  string
}

func New(p client.ConfigProvider, bucket string, keyPrefix string) *Backend {
	return &Backend{
		s3:         s3.New(p),
		uploader:   s3manager.NewUploader(p),
		downloader: s3manager.NewDownloader(p),
		bucket:     bucket,
		keyPrefix:  keyPrefix,
	}
}

func (b *Backend) CopyPlans(repoDir string, repoFullName string, env string, pullNum int) ([]plan.Plan, error) {
	// first list the plans with the correct prefix
	prefix := pathutil.Join(b.keyPrefix, repoFullName, strconv.Itoa(pullNum))
	list, err := b.s3.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(b.bucket), Prefix: &prefix})
	if err != nil {
		return nil, errors.Wrap(err, "listing plans")
	}

	var plans []plan.Plan
	for _, obj := range list.Contents {
		planName := pathutil.Base(*obj.Key)

		// only get plans from the correct env
		if planName != env+".tfplan" {
			continue
		}

		// determine the path relative to the repo
		relPath, err := filepath.Rel(prefix, *obj.Key)
		if err != nil {
			continue
		}
		downloadPath := filepath.Join(repoDir, relPath)
		file, err := os.Create(downloadPath)
		if err != nil {
			return nil, errors.Wrapf(err, "creating file %s to download plan to", downloadPath)
		}
		defer file.Close()

		_, err = b.downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(b.bucket),
				Key:    obj.Key,
			})
		if err != nil {
			return nil, errors.Wrapf(err, "downloading file at %s", *obj.Key)
		}
		plans = append(plans, plan.Plan{
			Project: models.Project{
				Path:         pathutil.Dir(relPath),
				RepoFullName: repoFullName,
			},
			LocalPath: downloadPath,
		})
	}
	return plans, nil
}

func (b *Backend) SavePlan(path string, project models.Project, env string, pullNum int) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "opening plan at %s", path)
	}

	key := b.path(project, env, pullNum)
	_, err = b.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(b.bucket),
		Key:    &key,
		Body:   f,
		Metadata: map[string]*string{
			"repoFullName": aws.String(project.RepoFullName),
			"path":         aws.String(project.Path),
			"env":          aws.String(env),
			"pullNum":      aws.String(strconv.Itoa(pullNum)),
		},
	})
	if err != nil {
		return errors.Wrap(err, "uploading plan to s3")
	}
	return nil
}

func (b *Backend) DeletePlan(project models.Project, env string, pullNum int) error {
	_, err := b.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.path(project, env, pullNum)),
	})
	return err
}

func (b *Backend) DeletePlansByPull(repoFullName string, pullNum int) error {
	// first list the plans with the correct prefix
	prefix := pathutil.Join(b.keyPrefix, repoFullName, strconv.Itoa(pullNum))
	list, err := b.s3.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(b.bucket), Prefix: &prefix})
	if err != nil {
		return errors.Wrap(err, "listing plans")
	}

	var deleteList []*s3.ObjectIdentifier
	for _, obj := range list.Contents {
		deleteList = append(deleteList, &s3.ObjectIdentifier{Key: obj.Key})
	}

	_, err = b.s3.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: aws.String(b.bucket),
		Delete: &s3.Delete{Objects: deleteList},
	})
	return err
}

func (b *Backend) path(project models.Project, env string, pullNum int) string {
	return pathutil.Join(b.keyPrefix, project.RepoFullName, strconv.Itoa(pullNum), project.Path, env+".tfplan")
}
