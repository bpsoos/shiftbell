dev_image := "shiftbell-dev"
export DEV_IMAGE := dev_image

default:
    just --list

build:
    docker compose build

shell: build
    docker run --rm -it -v {{justfile_dir()}}:/src {{dev_image}}

up: build
    docker compose up
