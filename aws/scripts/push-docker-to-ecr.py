import subprocess
import boto3
import os
import argparse

# Replace these with your own values
AWS_REGION = 'your-region'
REPOSITORY_NAME = 'my-ecr-repo'
IMAGE_NAME = 'your-image-name'
TAG = 'latest'

def run_command(command):
    """Run a shell command and check for errors."""
    try:
        subprocess.check_call(command, shell=True)
    except subprocess.CalledProcessError as e:
        print(f"An error occurred: {e}")
        exit(1)

def main(dockerfile_dir):
    # Initialize boto3 client for ECR
    ecr_client = boto3.client('ecr', region_name=AWS_REGION)
    
    # Create ECR repository if it doesn't exist
    try:
        response = ecr_client.create_repository(repositoryName=REPOSITORY_NAME)
        repository_uri = response['repository']['repositoryUri']
        print(f"Created repository {REPOSITORY_NAME}. URI: {repository_uri}")
    except ecr_client.exceptions.RepositoryAlreadyExistsException:
        print(f"Repository {REPOSITORY_NAME} already exists. Retrieving URI.")
        response = ecr_client.describe_repositories(repositoryNames=[REPOSITORY_NAME])
        repository_uri = response['repositories'][0]['repositoryUri']
    
    # Authenticate Docker to ECR
    print("Authenticating Docker to ECR...")
    login_command = f"aws ecr get-login-password --region {AWS_REGION} | docker login --username AWS --password-stdin {repository_uri.split('/')[0]}"
    run_command(login_command)
    
    # Navigate to the Dockerfile directory
    os.chdir(dockerfile_dir)
    print(f"Changed directory to {dockerfile_dir}")

    # Build Docker image
    print(f"Building Docker image {IMAGE_NAME} from Dockerfile in {dockerfile_dir}...")
    build_command = f"docker build -t {IMAGE_NAME} ."
    run_command(build_command)
    
    # Tag Docker image
    print(f"Tagging Docker image {IMAGE_NAME}...")
    tag_command = f"docker tag {IMAGE_NAME}:{TAG} {repository_uri}:{TAG}"
    run_command(tag_command)
    
    # Push Docker image to ECR
    print(f"Pushing Docker image {IMAGE_NAME} to ECR...")
    push_command = f"docker push {repository_uri}:{TAG}"
    run_command(push_command)

    print("Docker image pushed to ECR successfully!")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Build and push Docker image to AWS ECR.')
    parser.add_argument('--dockerfile-dir', required=True, help='Path to the directory containing the Dockerfile')
    
    args = parser.parse_args()
    
    main(args.dockerfile_dir)