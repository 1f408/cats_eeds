[cats\_dogs]: https://github.com/1f408/cats_dogs "Cats Dogs"
[cats\_pr\_dogs]: https://github.com/1f408/cats_pr_dogs "CatsPrDogs"
[cat\_mdview]: https://github.com/1f408/cats_dogs/blob/master/docs/cat_mdview.md
[cat\_tmplview]: https://github.com/1f408/cats_dogs/blob/master/docs/cat_tmplview.md
[user\_map形式]: https://github.com/1f408/cats_dogs/blob/master/docs/user_map.md

# cats\_eeds(CATs\_dogs SEEDS)

[cats\_dogs]の主な機能を提供しているコアライブラリです。  
派生物として[cats\_pr\_dogs](CatsPrDogs/cats\_dogs prviewerアプリ)にも使われています。

以下のモジュール群が含まれています。

- [view/mdview](../view/mdview) cats\_dogsの[cat\_mdview]のHTTPサーバハンドラ
- [view/tmplview](../view/tmplview) cats\_dogsの[cat\_tmplview]のHTTPサーバハンドラ
- [md2html](../md2html/) cats\_dogsのMarkdownをHTMLに変換する機能を扱うモジュール
- [authz](../authz/) cats\_dogsの[user\_map形式]ファイルでの認可処理ためのモジュール
- [upath](../upath/) cats\_dogsの内部パス形式(OSパスの互換性問題を回避)を扱うモジュール

利用方法については、[cats\_dogs]、[cats\_pr\_dogs]のコードを参考にしてください。
