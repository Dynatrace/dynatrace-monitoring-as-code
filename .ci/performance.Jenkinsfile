podTemplate(yaml: readTrusted('.ci/jenkins_agents/build-agent.yaml')) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"
        }
        stage("try GO") {
            container("monaco-build") {
              sh '''
                pwd
                ls -alF
                go version
              '''
            }
        }
    }
}
