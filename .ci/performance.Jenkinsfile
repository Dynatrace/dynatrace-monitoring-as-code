podTemplate(yaml: '''
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: monaco-build
      image: golang:latest
      imagePullPolicy: IfNotPresent
      resources:
        requests:
          cpu: "100m"
          memory: "1Gi"
        limits:
          cpu: "2"
          memory: "16Gi"
'''
) {
    node(POD_LABEL) {
        stage("HELLO") {
            echo "hello world"
        }
    }
}

