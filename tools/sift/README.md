# sift

行単位で入力を絞り込むコマンドです。複数の条件はすべて満たす必要があります。

```sh
go run ./sift -input words.txt -output result.txt range=2..8 charset=alpha
cat words.txt | go run ./sift initial=alpha
cat words.txt | go run ./sift -output - range=2..
go run ./sift -input addresses.txt -output - format=email 'match=*example.com'
go run ./sift -input words.txt initial=upper 'match=Ab*'
```

- `-input`: 入力ファイル。省略時または `-` は標準入力。
- `-output`: 出力ファイル。省略時はカレントディレクトリの `sift.txt`。`-` は標準出力。
- オプションは条件より前に指定します。`--input` / `--output` も利用できます。
- 出力ファイルが存在する場合は上書きします。入力と同じファイルは指定できません。
- `-h`: ヘルプ。
- 引数なしで端末などの文字デバイスから起動するとヘルプを表示します。パイプやファイルから入力する場合は、引数なしでも全行を処理します。
- 失敗時は標準エラー出力にエラーと対象コマンドのヘルプを表示し、終了コード1で終了します。

条件:

- `range=min..max`: Unicode の文字数による範囲（両端を含む）。片方の省略が可能。
- `initial=alpha|digit|alnum|upper`: 先頭文字の種類。空行には一致しません。
- `charset=alpha|digit|alnum|upper`: すべての文字の種類。
- `format=url|email|domain`: 組み込みの固定正規表現で行全体の形式を判定。
- `match=pattern`: 文字列との一致方法を、両端の `*` で指定。

`alpha` は Unicode の文字、`digit` は Unicode の数字、`alnum` はそのいずれかです。
`upper` は Unicode の大文字です。`initial=upper` で先頭が大文字の行を選べます。
条件を省略すると全行を出力します。入力は従来どおり `bufio.Scanner` の行長制限に従います。

## 正規表現の生成

通常は従来どおり入力をチェックして、一致した行を書き出します。
`regex` サブコマンドでは、同じ条件から PCRE2 形式の正規表現を生成します。

```sh
go run ./sift regex format=email
go run ./sift regex initial=upper 'match=Ab*'
go run ./sift regex -output pattern.txt range=2..8 charset=alpha
```

生成モードの出力は既定で標準出力です。`-output` / `--output` でファイルにも保存できます。
正規表現を1つ、改行付きで出力します。入力は読み込まず、`-input` は受け付けません。
オプションは条件より前に指定してください。不正な条件や生成できない条件では出力を上書きしません。

たとえば `regex initial=upper 'match=Ab*'` の出力は次のとおりです。

```text
\A(?=(?:\p{Lu}(?s:.*))\z)(?=(?:Ab(?s:.*))\z)(?s:.*)\z
```

先読みで各条件を AND として組み合わせ、文字列全体に一致させます。
条件なしの場合は、空文字列を含むすべての文字列に一致する式を生成します。
先読みを使うため、生成式は Go の `regexp` では使えません。PCRE2 エンジンを UTF モードで利用し、入力を行ごとに判定してください。
`rg --pcre2` は既定で Unicode を扱うため、そのまま利用できます。
正規表現には区切り文字や引用符を付けません。利用先に応じて囲んでください。
`range` の上限・下限が65535を超える場合は PCRE2 の繰り返し回数の制約で生成エラーになります。通常のチェックにはこの制約はありません。

## 形式の判定

すべて行全体に一致させます。前後の空白は取り除きません。

| 条件 | 対象と例 |
| --- | --- |
| `format=url` | HTTP/HTTPS URL。`https://example.com/path?q=1`。ASCII のホスト名、IPv4 表記、`localhost`、任意の数字1〜5桁のポートとパス・クエリ・フラグメントを受け付けます。 |
| `format=email` | ASCII のメールアドレス。`first.last+tag@example.com`。ローカル部の先頭・末尾・連続したドットは受け付けません。 |
| `format=domain` | ドットで区切られた ASCII ドメイン名。`sub.example.com`。各ラベルは1〜63文字で、英数字とハイフンを使用できます。先頭・末尾のハイフンは不可で、最終ラベルは英字で始めます。 |

メールのドメイン部分も `format=domain` と同じ規則です。
国際化ドメインは Punycode 表記を使用してください。末尾のドットは受け付けません。
URL のユーザー情報と IPv6 表記、メールの引用符付きローカル部は対象外です。
固定の正規表現による形式の判定であり、ポート・IPアドレスの数値範囲、ドメイン全体の長さ、接続先やメールアドレスの実在までは確認しません。

## 文字列の一致

| 条件 | 一致方法 | 一致する例 |
| --- | --- | --- |
| `match=abc` | 完全一致 | `abc` |
| `match=abc*` | 前方一致 | `abcdef` |
| `match=*abc` | 後方一致 | `xyzabc` |
| `match=*abc*` | 部分一致 | `xyzabcdef` |

大文字・小文字を区別し、正規表現の記号も通常の文字として比較します。
`*` は両端に一つずつ指定でき、途中の `*` や連続した `*` はエラーです。
ただし `match=*` と `match=**` は全行に一致します。`match=` は空行だけに一致します。
文字としての `*` を指定するエスケープはありません。
シェルによる展開を防ぐため、`*` を含む条件は `'match=*abc*'` のように引用符で囲んでください。
複数の `match` も、ほかの条件と同様に AND で評価します。

構成:

- `main.go`: コマンドの起動と終了コード。
- `internal/cli`: オプション、ファイルの開閉、入出力先の指定。
- `internal/filter/conditions.go`: 条件の解釈・検証と評価順の決定。
- `internal/filter/formats.go`: 固定正規表現による形式の判定。
- `internal/filter/text.go`: 文字列一致条件の準備と判定。
- `internal/filter/regex.go`: 準備した条件から PCRE2 正規表現を生成。
- `internal/filter/filter.go`: 条件の評価と行の読み書き。
- 各パッケージの `errors.go`: エラー型。`Unwrap` によって `errors.Is` / `errors.As` に対応。
