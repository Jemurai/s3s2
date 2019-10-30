pipeline {
    agent {
        docker {
            image 'golang'
        }
    }
    stages {
        stage('build') {
            steps {
                sh script: "chmod u+x ./deploy_build.sh", label: "Permissioning build file..."
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
            }
            steps {
                // Fun stuff it **** anything similar to NEXUS_CREDS
                sh 'echo $NEXUS_CREDS_USR should be leeroy-tempus-n if this worked :pray:'
                sh '''
                curl --fail --user "${NEXUS_CREDS}" \
                    --upload-file ./where-is-the-file-i-dont-know \
                    https://nexus.securetempus.com/repository/tempus-n/literally-any-path/i-dont-care/havefun
                '''
            }
        }
    }
}
