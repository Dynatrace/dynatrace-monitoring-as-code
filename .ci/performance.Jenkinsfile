podTemplate(yamlFile: '.ci/jenkins_agents/ca-jenkins-agent.yaml'
) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"

            def maven = docker.image('maven:latest')
            maven.pull()
        }
    }
}

