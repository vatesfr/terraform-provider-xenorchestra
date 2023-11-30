pipeline {
  // Run on an agent where we want to use Go
  agent any

  // Ensure the desired Go version is installed for all stages,
  // using the name defined in the Global Tool Configuration
  tools { go '1.20' }

  stages {
    stage('BuildAndTest') {
    matrix {

      stages {
        stage('Test') {

          steps {
            // Output will be something like "go version go1.19 darwin/arm64"
            sh 'go version'
            sh 'cp /opt/terraform-provider-xenorchestra/testdata/images/alpine-virt-3.17.0-x86_64.iso xoa/testdata/alpine-virt-3.17.0-x86_64.iso'
            sh 'make ci'
          }
        }
      }
    }
  }
}
