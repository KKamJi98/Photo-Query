pipeline {
    agent any
    
    tools {
        go 'kkam_go'
    }
    environment {
        GITNAME = 'war-oxi'                 // 본인 깃허브계정
        GITMAIL = 'xowl5460@naver.com'      // 본인 이메일
        GITWEBADD = 'https://github.com/War-Oxi/ACE-Team-KKamJi.git'
        GITSSHADD = 'git@github.com:War-Oxi/ACE-Team-KKamJi.git'
        GITCREDENTIAL = 'kkam_git_cre'           // 아까 젠킨스 credential에서 생성한

        ECR_URL = '109412806537.dkr.ecr.us-east-1.amazonaws.com/app-picture-backend'
        ECR_CREDENTIAL = 'aws_cre'
    }
    
    stages {
        stage('Checkout Github') {
            steps {
                checkout([$class: 'GitSCM', branches: [[name: '*/main']], extensions: [],
                userRemoteConfigs: [[credentialsId: GITCREDENTIAL, url: GITWEBADD]]])
            }
            post {
                failure {
                    echo 'Repository clone failure'
                }
                success {
                    echo 'Repository clone success'
                }
            }
        }
        #docker image build
        stage('image build') {
            steps {
                sh "docker build -t ${ECR_URL}:${currentBuild.number} ."
                sh "docker build -t ${ECR_URL}:latest ."
            }
        }
        stage('image push') {
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY', credentialsId: ECR_CREDENTIAL]]) {
                    script {
                        def ecrLogin = sh(script: "aws ecr get-login-password --region us-east-1", returnStdout: true).trim()
                        sh "docker login -u AWS -p ${ecrLogin} ${ECR_REPO_URL}"
                    }
                }
                // docker image ECR로 Push
                sh "docker push -t ${ECR_URL}:${currentBuild.number} ."
                sh "docker build -t ${ECR_URL}:latest ."
            }

            post {
                failure {
                    echo 'AWS ECR로 이미지 푸시 실패'
                    sh "docker image rm -f ${ECR_REPO_URL}:${currentBuild.number}"
                    sh "docker image rm -f ${ECR_REPO_URL}:latest"
                }
                
                success {
                    echo 'AWS ECR로 이미지 푸시 성공'
                    sh "docker image rm -f ${ECR_REPO_URL}:${currentBuild.number}"
                    sh "docker image rm -f ${ECR_REPO_URL}:latest"
                }
            }
        }
    }
}