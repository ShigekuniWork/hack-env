_default:
    @just --list

setup-wordlists:
    mkdir -p wordlists
    test -d wordlists/SecLists || \
        git clone --depth 1 https://github.com/danielmiessler/SecLists.git wordlists/SecLists

update-wordlists:
    git -C wordlists/SecLists pull --ff-only