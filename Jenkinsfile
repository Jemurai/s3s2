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
                sh 'echo $NEXUS_CREDS_USR should be leeroy-tempus-n if this worked :pray:'
            }
        }
    }
}
