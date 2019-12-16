pipeline {
    agent {
        // pipeline-model-definition (might be the correct plugin?) is not
        // trying to pull from dockerhub if the image does not exist locally on
        // the jenkins agent. By default it was falling back to a docker
        // registry at the same url as the jenkins master, which does not and
        // will not ever exist. To address, we are explicitly setting the
        // registryUrl to the dockerhub v2 api.
        docker {
            image 'golang'
            registryUrl "https://index.docker.io/v2/"
            args '--user root'
        }
    }
    stages {
        stage('build') {
            steps {
                sh script: "./deploy_build.sh", label: "Building..."
                sh script: 'echo Built successfully!', label: "Build successful!"
                }
        }
        stage('publish') {
            environment {
                // credentials() will create three environment variables
                // NEXUS_CREDS = username:password
                // NEXUS_CREDS_USR = username
                // NEXUS_CREDS_PSW = password
                // https://jenkins.io/doc/book/pipeline/jenkinsfile/#handling-credentials
                NEXUS_CREDS = credentials('nexus-leeroy-tempus-n')
                NEXUS_PATH = 'https://nexus.securetempus.com/repository/tempus-n'
                S3S2_VERSION = "${GIT_COMMIT}"
                PUBLIC_S3_BUCKET = 'tdo-n-message-gateway-s3s2-use1'
            }
            steps {
                // Fun stuff it will mask out with **** anything similar to NEXUS_CREDS
                // ${GIT_COMMIT} is the commit hash if you want to use that
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./linux/s3s2-linux-amd64 ${NEXUS_PATH}/${S3S2_VERSION}/s3s2-linux-amd64', label: "Publishing Linux build"
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./darwin/s3s2-darwin-amd64 ${NEXUS_PATH}/${S3S2_VERSION}/s3s2-darwin-amd64', label: "Publishing Mac build"
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./windows/s3s2-windows-amd64.exe ${NEXUS_PATH}/${S3S2_VERSION}/s3s2-windows-amd64.exe', label: "Publishing Windows build"
                sh script: 'apt update && apt install -y python-pip && pip --no-cache-dir install awscli'
                sh script: 'aws s3api put-object --bucket ${PUBLIC_S3_BUCKET} --key ${S3S2_VERSION}/s3s2-linux-amd64 --body ./linux/s3s2-linux-amd64 --acl public-read', label: "Publishing Linux build to S3"
                sh script: 'aws s3api put-object --bucket ${PUBLIC_S3_BUCKET} --key ${S3S2_VERSION}/s3s2-darwin-amd64 --body ./darwin/s3s2-darwin-amd64 --acl public-read', label: "Publishing darwin build to S3"
                sh script: 'aws s3api put-object --bucket ${PUBLIC_S3_BUCKET} --key ${S3S2_VERSION}/s3s2-windows-amd64.exe --body ./windows/s3s2-windows-amd64.exe --acl public-read', label: "Publishing windows build to S3"
            }
        }
    }
}
