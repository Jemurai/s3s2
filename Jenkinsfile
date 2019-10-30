pipeline {
    agent {
        docker {
            image 'golang'
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
            }
            steps {
                // Fun stuff it will mask out with **** anything similar to NEXUS_CREDS
                // ${GIT_COMMIT} is the commit hash if you want to use that
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./linux/s3s2-linux-amd64 ${NEXUS_PATH}/s3s2-linux-amd64', label: "Publishing Linux build"
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./darwin/s3s2-darwin-amd64 ${NEXUS_PATH}/s3s2-darwin-amd64', label: "Publishing Mac build"
                sh script: 'curl --fail --user "${NEXUS_CREDS}" --upload-file ./windows/s3s2-windows-amd64.exe ${NEXUS_PATH}/s3s2-windows-amd64.exe', label: "Publishing Windows build"
            }
        }
    }
}
