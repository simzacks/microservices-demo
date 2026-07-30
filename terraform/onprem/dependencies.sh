go install sigs.k8s.io/cloud-provider-kind@latest
sudo install ~/go/bin/cloud-provider-kind /usr/local/bin

# to allow the service to work for a non-root user
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/cloud-provider-kind
sudo dnf install yq -y
sudo dnf install trivy -y
curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash

