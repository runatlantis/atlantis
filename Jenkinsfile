/*
 * Copyright 2018-present Sonatype Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
@Library(['private-pipeline-library', 'jenkins-shared']) _

dockerizedBuildPipeline(
  buildImageId: "${sonatypeDockerRegistryId()}/cdi/golang-1.14:1",
  prepare: {
    githubStatusUpdate('pending')
  },
  buildAndTest: {
    sh '''
    go mod download
    go mod tidy
    go get -u github.com/jstemmer/go-junit-report
    go test ./... -v 2>&1 -p=1 | go-junit-report > test-results.xml
    CGO_ENABLED=0 GOOS=linux go build -o nancy .
    '''
  },
  vulnerabilityScan: {
    withDockerImage(env.DOCKER_IMAGE_ID, {
      withCredentials([usernamePassword(credentialsId: 'policy.s integration account',
        usernameVariable: 'IQ_USERNAME', passwordVariable: 'IQ_PASSWORD')]) {
        sh 'go list -json -m all | ./nancy iq --iq-application nancy --iq-stage build --iq-username $IQ_USERNAME --iq-token $IQ_PASSWORD --iq-server-url https://policy.ci.sonatype.dev'
      }
    })
  },
  testResults: [ 'test-results.xml' ],
  onSuccess: {
    githubStatusUpdate('success')
  },
  onFailure: {
    githubStatusUpdate('failure')
    //notifyChat(currentBuild: currentBuild, env: env, room: 'community-oss-fun')
   //sendEmailNotification(currentBuild, env, [], 'someone@sonatype.com')
  }
)
