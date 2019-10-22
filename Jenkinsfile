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
            steps {
                sh 'echo Pretend I published'
            }
        }
    }
}
