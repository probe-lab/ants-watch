# Prints available recipes
default:
    @just --list --justfile {{ justfile() }}

repo := env_var_or_default("REPO", "019120760881.dkr.ecr.us-east-1.amazonaws.com")
repo_user := env_var_or_default("REPO_USER", "AWS")
repo_region := env_var_or_default("REPO_REGION", "us-east-1")
tag := env_var_or_default("TAG", `date +%F` + "-" + `echo -n "${ANTS_REPO:-github.com/probe-lab/ashby}-${ANTS_COMMIT:-839c2f242c67825ff7884515e929c7ba5b16e84e}" | sha256sum | cut -c -8` + "-" + `git describe --always --tag --dirty`)
image_name := env_var_or_default("IMAGE_NAME", "probelab:ants-" + tag)

# Format, vet, and lint the codebase
check:
    gofmt -w .
    go vet ./...
    golangci-lint run

# Build the linux/amd64 docker image
build:
    @echo "{{ tag }}"
    docker build --platform="linux/amd64" -t "{{ image_name }}" .

# Push the docker image to the ECR registry
push:
    aws ecr get-login-password --region {{ repo_region }} | docker login --username {{ repo_user }} --password-stdin {{ repo }}
    docker tag "{{ image_name }}" "{{ repo }}/{{ image_name }}"
    docker push "{{ repo }}/{{ image_name }}"
