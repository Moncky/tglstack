# tglstack
Prepare Terraform and Terragrunt configs for LocalStack

This tool will scan down a terraform directory structure and move any provider and terragrunt blocks to a single file that can later be overridden in terragrunt with the `generate` function

At the moment it is a broken mess because its the fist thing I have written in golang and I have no idea what I am doing  Don't use for anything you don't want to potentially lose