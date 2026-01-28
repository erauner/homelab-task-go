@Library('homelab-jenkins-library@main') _

pipeline {
    agent {
        kubernetes {
            yaml homelab.podTemplate('kaniko-go')
        }
    }

    environment {
        REGISTRY = 'docker.nexus.erauner.dev'
        IMAGE_NAME = "homelab/${env.JOB_BASE_NAME.replaceAll('/', '-')}"
    }

    stages {
        stage('Test') {
            steps {
                container('golang') {
                    sh 'go test ./... -v -coverprofile=coverage.out'
                }
            }
            post {
                always {
                    junit allowEmptyResults: true, testResults: '**/junit.xml'
                }
            }
        }

        stage('Build & Push Image') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag pattern: 'v*', comparator: 'GLOB'
                }
            }
            steps {
                container('kaniko') {
                    script {
                        def imageTag = env.TAG_NAME ?: 'latest'
                        def gitCommit = homelab.gitShortCommit()
                        sh """
                            /kaniko/executor \
                                --destination=${REGISTRY}/${IMAGE_NAME}:${imageTag} \
                                --destination=${REGISTRY}/${IMAGE_NAME}:${gitCommit} \
                                --cache=true \
                                --cache-repo=${REGISTRY}/${IMAGE_NAME}/cache
                        """
                    }
                }
            }
        }
    }

    post {
        failure {
            script {
                homelab.notifyDiscord(status: 'FAILURE')
            }
        }
    }
}
