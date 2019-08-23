pipeline {
    agent {
        docker {
            image 'golang:1.12.9-alpine3.10'
        }
    }
    stages {
        stage('build') {
            steps {
                sh 'make build'
            }
        }
        stage('publish') {
            steps {
                sh 'echo Pretend I published'
            }
        }
    }
}
