podTemplate(yaml: readTrusted('.ci/jenkins_agents/build-agent.yaml')) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"
            sh '''
                pwd
            '''
        }
        stage("try GO") {
            container("monaco-build") {
                checkout scm
                echo "done"
                dir('a-child-repo') {
                    sh 'pwd'
                    git credentialsId: 'bitbucket-buildmaster',
                        url: 'https://bitbucket.lab.dynatrace.org/scm/claus/monaco-test-data.git',
                        branch: 'main'
                        sh '''
                            pwd
                            ls -alF
                        '''
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
