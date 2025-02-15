# Kino
Totus mundus agit histrionem.


## Architecture

This consists of two parts

1. Scraper
    - This goes through a list of providers, flixhq at the moment, to pull movie information.
2. CLI
    - Acts as the model to control everything. Allows for the viewing of images on the terminal if you have that capability.
    Also has caching for movie images.
3. Decrypter
    - This is a seperate portion, that should be run with docker
    this will open a port that allows for the embeded ID's to be decrypted locally.
    ```yaml
        docker pull eatmynerds/embed_decrypt
        docker run -p 3000:3000 eatmynerds/embed_decrypt
    ```
    Adding the -d flag allows it to be run as a detached process (just remember to kill the process after)

