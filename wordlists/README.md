# Wordlist

ブルートフォースや、サブドメインの列挙などで使用する単語リスト。

## ツール

### CUPP

人物情報からパスワード候補を生成するツール

```shell
just wordlists gen-passowrd-list
```

### username-anarchy

氏名からありがちなusername候補を生成するツール

```shell
just wordlists gen-username-list <firstname> <lastname>　<アウトプット先>
```

## 外部のワードリスト

外部のワードリストをクローンして使用します。

### 環境事項築
```shell
just setup
```

### SecLists

有名なワードリストのリポジトリ

[SecLists](https://github.com/danielmiessler/seclists)