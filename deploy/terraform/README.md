# How to use Terraform config

You will need a valid `kubeconfig` to point to in the `provider.tf` file, and sufficient permissions to deploy this to your cluster. This version is based off the `install.yaml` file at the time of writing and would need to be updated in lock step to reflect any changes there.

There are variables for the name  and the namespace that it deploys to, are just to make it more convenient to change them if you wish. The arguments being in a variable would allow you to (for example) have them as an easily editable variable in Terraform Cloud/Enterprise and change them without making any changes to the Terraform code itself and similarly the version tag of the container has been made a variable to make updates easier to deploy.