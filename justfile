_default:
    @just --list

# Authのjustfile
auth_justfile := "auth/justfile"
# WordListsのjustfile
wordlists_justfile := "wordlists/justfile"

# 環境のセットアップを行う
[group('setup')]
setup:
    just --justfile {{wordlists_justfile}} setup

# 環境のアップデート
[group('setup')]
update:
    just --justfile {{wordlists_justfile}} update

# 認証関連のコマンドを使用する
[group('command')]
auth *args:
    just --justfile {{auth_justfile}} {{args}}

# ワードリスト関連のコマンドを使用する
[group('command')]
wordlists *args:
    just --justfile {{wordlists_justfile}} {{args}}
