pipeline {
  agent any
  environment {
    XOA_URL               = credentials("terraform-provider-xoa-url")
    XOA_USER              = credentials("terraform-provider-xoa-user")
    XOA_PASSWORD          = credentials("terraform-provider-xoa-password")
    XOA_POOL              = credentials("terraform-provider-xoa-pool")
    XOA_TEMPLATE          = credentials("terraform-provider-xoa-template")
    XOA_DISKLESS_TEMPLATE = credentials("terraform-provider-xoa-diskless-template")
    XOA_ISO               = credentials("terraform-provider-xoa-iso")
    XOA_ISO_SR            = credentials("terraform-provider-xoa-iso-sr")
    XOA_NETWORK           = credentials("terraform-provider-xoa-network")
    XOA_RETRY_MAX_TIME    = credentials("terraform-provider-xoa-retry-max-time")
    XOA_RETRY_MODE        = credentials("terraform-provider-xoa-retry-mode")
  }

  // Ensure the desired Go version is installed for all stages,
  // using the name defined in the Global Tool Configuration
  tools { go '1.21' }

  stages {
    stage('BuildAndTest') {
    matrix {
      axes {
        axis {
          name 'TF_VERSION'
          values 'terraform-v0.14.11', 'terraform-v1.7.0'
        }
      }
      stages {
        stage('Test') {
          environment {
            BYPASS_XOA_TOKEN = sh(script: "xo-cli --createToken $XOA_URL $XOA_USER $XOA_PASSWORD | tail -n1", returnStdout: true).trim()
          }
          steps {
            lock('xoa-test-runner') {
              sh 'cp /opt/terraform-provider-xenorchestra/testdata/images/alpine-virt-3.17.0-x86_64.iso xoa/testdata/alpine-virt-3.17.0-x86_64.iso'
              sh 'TF_VERSION=${TF_VERSION} TIMEOUT=60m make ci'
            }
          }
        }
      }
    }
    }
  }
}
