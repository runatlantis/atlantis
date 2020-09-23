/*
 * Copyright (c) 2011-present Sonatype, Inc. All rights reserved.
 * Includes the third-party code listed at http://links.sonatype.com/products/clm/attributions.
 * "Sonatype" is a trademark of Sonatype, Inc.
 */

@Library(['private-pipeline-library', 'jenkins-shared']) _

def workDir = "./"

dockerizedBuildPipeline(
  buildImageId: "${sonatypeDockerRegistryId()}/cdi/golang-1.14:1",
  prepare: {
    githubStatusUpdate('pending')
  },
  buildAndTest: {
    dir(workDir) {
      runSafely '''
      go mod download
      GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o atlantis
      '''
    }
  },
  vulnerabilityScan: {
    nancyEvaluation(workDir + '/go.sum')
  },
  // deploy: {
  //   dir(workDir) {
  //     def region = 'us-east-2'
  //     withAWS(role: config.role, roleAccount: config.account, region: region) {
  //       def odsPurgeName = "hds-${params.ENVIRONMENT}-ods-purge-${region}"
  //       runSafely 'zip ods-purge.zip ods-purge'
  //       s3Upload(acl: 'Private', bucket: odsPurgeName, sseAlgorithm:'AES256', file:'ods-purge.zip')
  //       runSafely "aws lambda update-function-code --function-name ${odsPurgeName} --s3-bucket ${odsPurgeName} --s3-key ods-purge.zip"
  //     }
  //   }
  // },
  onSuccess: {
    githubStatusUpdate('success')
  },
  onFailure: {
    githubStatusUpdate('failure')
  }
)
