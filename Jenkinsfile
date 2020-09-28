/*
 * Copyright (c) 2011-present Sonatype, Inc. All rights reserved.
 * Includes the third-party code listed at http://links.sonatype.com/products/clm/attributions.
 * "Sonatype" is a trademark of Sonatype, Inc.
 */

@Library(['private-pipeline-library', 'jenkins-shared']) _


dockerizedBuildPipeline(
  pathToDockerfile: 'jenkins/Dockerfile',
  prepare: {
    githubStatusUpdate('pending')
  },
  buildAndTest: {
    // Following instructions for building and testing from 
    // https://github.com/runatlantis/atlantis/blob/master/CONTRIBUTING.md#running-atlantis-locally
    runSafely '''
    go install
    make test
    '''
  },
  vulnerabilityScan: {
    nexusPolicyEvaluation iqApplication: 'atlantis', iqScanPatterns: [[scanPattern: '**/Gopkg.lock'], [scanPattern: '**/go.sum'], [scanPattern: '**/go.list']], iqStage: 'build'
    //nancyEvaluation(workDir + '/go.sum')
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
//  testResults: [ 'test-results.xml' ],
  onSuccess: {
    githubStatusUpdate('success')
  },
  onFailure: {
    githubStatusUpdate('failure')
  }
)
