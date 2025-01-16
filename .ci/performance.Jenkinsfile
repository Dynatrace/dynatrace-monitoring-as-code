podTemplate(yaml: readTrusted('.ci/jenkins_agents/build-agent.yaml')) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"
            sh '''
                apt install make
            '''
        }
        stage("🏗️ building") {
            container("monaco-build") {
                echo "done"
                dir('/tmp/source') {
                    checkout scm
//                     git credentialsId: 'bitbucket-buildmaster',
//                         url: 'https://bitbucket.lab.dynatrace.org/scm/claus/monaco-test-data.git',
//                         branch: 'main'
                    sh '''
                        pwd
                        ls -alF
                    '''
                    deleteDir()
                }
                sh '''
                    pwd
                    ls -alF
                    go version
                '''
            }
        }
    }
}
