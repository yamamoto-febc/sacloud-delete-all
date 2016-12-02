# sacloud-delete-all

さくらのクラウド(IaaS)上のリソースをすべて削除するコマンド

## インストール

[リリースページ](https://github.com/yamamoto-febc/sacloud-delete-all/releases/latest)にて実行ファイルをzipで配布しています。

ダウンロードして展開、実行権を付与してください。

## 使い方

```bash
$ ./sacloud-delete-all --token=[さくらのクラウドAPIトークン] --secret=[さくらのクラウドAPIシークレット]

# 実行すると以下の確認が表示される。
Do you really want to destroy all?[Y/n]

# Yを入力すると削除実行

```

指定できるオプションの詳細は`--help`で表示できます。

```bash
$ ./sacloud-delete-all --help

NAME:
   sacloud-delete-all - A CLI tool of to delete all resources on Sakura Cloud

USAGE:
   sacloud-delete-all [options]

REQUIRED PARAMETERS:
   --token value, --sakuracloud-access-token value          API Token of SakuraCloud (default: none) [$SAKURACLOUD_ACCESS_TOKEN]
   --secret value, --sakuracloud-access-token-secret value  API Secret of SakuraCloud (default: none) [$SAKURACLOUD_ACCESS_TOKEN_SECRET]
   
OPTIONS:
   --zones value, --sakuracloud-zones value  Target zone list of SakuraCloud (default: "tk1v", "is1a", "is1b", "tk1a") [$SAKURACLOUD_ZONES]
   --sakuracloud-trace-mode                  Flag of SakuraCloud debug-mode (default: false) [$SAKURACLOUD_TRACE_MODE]
   --force                                   Flag of force delete mode (default: false) [$FORCE]
   --trace-log                               Flag of enable TRACE log (default: false) [$TRACE_LOG]
   --info-log                                Flag of enable INFO log (default: true) [$INFO_LOG]
   --warn-log                                Flag of enable WARN log (default: true) [$WARN_LOG]
   --error-log                               Flag of enable ERROR log (default: true) [$ERROR_LOG]
   --help, -h                                show help (default: false)
   --version, -v                             print the version (default: false)
   
VERSION:
   0.0.1, build xxxxxxxx

```

## 注意点

- `--zones`オプションで一部ゾーンのみを対象とした場合、かつ削除対象にブリッジが含まれる場合、ブリッジの削除が行えない場合があります。

- ブリッジにおいて、専用サーバー/VPSスイッチと接続されている場合はブリッジの削除が行えません。

## 対象リソース

以下のリソースが削除されます。

- サーバー
- ディスク
- アーカイブ
- 自動バックアップ
- ISOイメージ
- スイッチ
- スイッチ+ルーター
- パケットフィルタ
- ブリッジ
- ロードバランサ
- VPCルータ
- データーベース
- GSLB
- DNS
- シンプル監視
- ライセンス
- 公開鍵
- スクリプト
- アイコン

以下のリソースはAPI経由で操作できないため削除されません。

- 割引パスポート
- APIキー
- クーポン

## License

 `sacloud-delete-all` Copyright (C) 2016 Kazumichi Yamamoto.

  This project is published under [Apache 2.0 License](LICENSE.txt).
  
## Author

  * Kazumichi Yamamoto ([@yamamoto-febc](https://github.com/yamamoto-febc))

