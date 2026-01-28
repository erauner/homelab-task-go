#!/usr/bin/env groovy
/**
 * Jenkinsfile for homelab-task-go
 *
 * Builds and pushes the Go task toolkit container image on every merge to main.
 * Uses semantic versioning based on git tags.
 *
 * Image: docker.nexus.erauner.dev/homelab/taskkit-go:<version>
 */

@Library('homelab') _

// Inline pod template with Go + Kaniko
def podYaml = '''
apiVersion: v1
kind: Pod
metadata:
  labels:
    workload-type: ci-builds
spec:
  imagePullSecrets:
  - name: nexus-registry-credentials
  containers:
  - name: jnlp
    image: jenkins/inbound-agent:3355.v388858a_47b_33-3-jdk21
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
  - name: golang
    image: golang:1.22-alpine
    command: ['cat']
    tty: true
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 1000m
        memory: 1Gi
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    command: ['sleep', '3600']
    volumeMounts:
    - name: nexus-creds
      mountPath: /kaniko/.docker
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: 1000m
        memory: 2Gi
  volumes:
  - name: nexus-creds
    secret:
      secretName: nexus-registry-credentials
      items:
      - key: config.json
        path: config.json
'''

pipeline {
    agent {
        kubernetes {
            yaml podYaml
        }
    }

    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 15, unit: 'MINUTES')
        disableConcurrentBuilds()
    }

    environment {
        IMAGE_NAME = 'docker.nexus.erauner.dev/homelab/taskkit-go'
    }

    stages {
        stage('Setup') {
            steps {
                container('golang') {
                    sh '''
                        echo "=== Go version ==="
                        go version

                        echo "=== Downloading dependencies ==="
                        go mod download

                        echo "=== Verifying modules ==="
                        go mod verify
                    '''
                }
            }
        }

        stage('Lint') {
            steps {
                container('golang') {
                    sh '''
                        echo "=== Running go vet ==="
                        go vet ./...

                        echo "=== Checking formatting ==="
                        gofmt -l . | tee /tmp/gofmt.out
                        if [ -s /tmp/gofmt.out ]; then
                            echo "ERROR: Some files need formatting:"
                            cat /tmp/gofmt.out
                            exit 1
                        fi
                        echo "All files properly formatted"
                    '''
                }
            }
        }

        stage('Test') {
            steps {
                container('golang') {
                    sh '''
                        echo "=== Running tests ==="
                        go test ./... -v -coverprofile=coverage.out
                    '''
                }
            }
        }

        stage('Build Check') {
            steps {
                container('golang') {
                    sh '''
                        echo "=== Building binary ==="
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o taskkit ./cmd/taskkit

                        echo "=== Testing CLI entrypoint ==="
                        ./taskkit --help
                        ./taskkit list-handlers

                        echo "=== Binary size ==="
                        ls -lh taskkit
                    '''
                }
            }
        }

        stage('Build and Push Image') {
            steps {
                script {
                    env.VERSION = homelab.gitDescribe()
                    env.COMMIT = homelab.gitShortCommit()
                    echo "Building image version: ${env.VERSION} (commit: ${env.COMMIT})"

                    // Build and push using shared library (versioned tag)
                    homelab.homelabBuild([
                        image: env.IMAGE_NAME,
                        version: env.VERSION,
                        commit: env.COMMIT,
                        dockerfile: 'Dockerfile',
                        context: '.'
                    ])
                }
            }
        }

        stage('Push Latest Tag') {
            steps {
                container('kaniko') {
                    sh """
                        /kaniko/executor \
                            --dockerfile=Dockerfile \
                            --context=dir://. \
                            --destination=${IMAGE_NAME}:latest \
                            --cache=true \
                            --cache-repo=${IMAGE_NAME}/cache \
                            --custom-platform=linux/amd64
                    """
                }
            }
        }
    }

    post {
        success {
            echo """
            ✅ Build successful!

            Docker Image: ${env.IMAGE_NAME}:${env.VERSION}
            Latest Tag:   ${env.IMAGE_NAME}:latest

            To pull: docker pull ${env.IMAGE_NAME}:latest
            """
        }
        failure {
            script {
                homelab.postFailurePrComment([repo: 'erauner/homelab-task-go'])
                homelab.notifyDiscordFailure()
            }
        }
    }
}
