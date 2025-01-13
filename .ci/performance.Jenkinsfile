podTemplate(yaml: readTrusted('.ci/jenkins_agents/build-agent.yaml')) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"
        }
        stage("try GO") {
            container("monaco-build") {
               checkout scm
//                 dir('a-child-repo') {
//                     git branch: 'main', url: 'https://bitbucket.lab.dynatrace.org/scm/claus/monaco-test-data.git'
//                 }
                sh '''
                pwd
                ls -alF
                go version
              '''
            }
        }
    }
}
