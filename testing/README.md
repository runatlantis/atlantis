## Test Docker Image

The Docker image used by the tests is `runatlantis/testing-env`. It's built by the Dockerfie
in this directory.

The image is automatically built by Docker Hub on merges to `master`.

To update the image, create a pull request that updates the Dockerfile and get it merged to `master`.

Wait until the build at https://hub.docker.com/repository/docker/runatlantis/testing-env/builds finishes
and then scroll to the bottom to find output like:

```
01216b26cd75183360909a50217a39d55a9265e7: digest: sha256:9c26943a576bf8aaa7a3790f3f8677c68747114e027cfbc361717f49b958e2d1 size: 4741
Build finished
```

The sha `01216b26cd75183360909a50217a39d55a9265e7` is the tag of the latest image.

In `.circleci/config.yml`, update the image to reference the new tag:

```diff
jobs:
  test:
    docker:
-    - image: runatlantis/testing-env:<old tag>
+    - image: runatlantis/testing-env:<new tag>
```

And open up a PR.
