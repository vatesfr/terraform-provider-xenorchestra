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
  tools { go '1.20' }

  stages {
    stage('BuildAndTest') {
    matrix {
      axes {
        axis {
          name 'TF_VERSION'
          values 'terraform-v0.14.11', 'terraform-v1.6.6'
        }
      }
      stages {
        stage('Test') {
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
