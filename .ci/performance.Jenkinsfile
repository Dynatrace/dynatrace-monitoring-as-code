pipeline {
    agent {
        kubernetes {
            cloud 'linux-amd64'
            nodeSelector 'kubernetes.io/arch=amd64,kubernetes.io/os=linux'
            instanceCap '2'
            yamlFile '.ci/jenkins_agents/ca-jenkins-agent.yaml'
        }
    }

    stages {

        stage("HELLO") {
            steps {
                echo "hello world"
            }
        }
    }
}

