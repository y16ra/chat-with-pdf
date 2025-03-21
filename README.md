# Chat with PDF

ChatPDF APIを使用して、PDFファイルの内容についてチャットするCLIアプリケーションです。

## 機能

- PDFファイルのアップロード（ローカルファイルまたはURL）
- PDFの内容に関する質問応答
- 会話の履歴を保持したチャット機能

## 必要条件

- Go 1.22以上
- ChatPDF APIキー

## インストール

```bash
git clone https://github.com/y16ra/chat-with-pdf.git
cd chat-with-pdf
go build
```

## 使用方法

1. 環境変数の設定（オプション）:
```bash
export CHATPDF_API_KEY="your-api-key"
```

2. アプリケーションの実行:
```bash
./chat-with-pdf
```

3. メニューから操作を選択:
   - PDFファイルのアップロード
   - URLからPDFを追加
   - 既存のSourceIDを使用してチャットを開始

## 注意事項

- PDFファイルは最大2,000ページまたは32MBまでの制限があります
- APIキーは必須です（環境変数で設定するか、プログラム実行時に入力）
