dev_image := "shiftbell-dev"

default:
    just --list

build:
    docker build -t {{dev_image}} .

shell: build
    docker run --rm -it -v {{justfile_dir()}}:/src {{dev_image}}

run: build
    docker run --rm -it \
        -p 8080:80 \
        -v {{justfile_dir()}}:/src \
        {{dev_image}} \
        go tool air
