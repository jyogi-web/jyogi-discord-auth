# Rails: データベース直接参照 (Multiple Databases)

Ruby on Rails 6.0以上で導入された [Multiple Databases（複数データベース）](https://railsguides.jp/active_record_multiple_databases.html) 機能を使用して、じょぎ認証システムのデータベース（`users` テーブルなど）を直接参照する方法を解説します。

> [!WARNING]
> **推奨事項**: 基本的には [OAuth2 API連携](./client-integration.md) を使用することを強く推奨します。
> データベースを直接参照する場合、認証サーバーのスキーマ変更（マイグレーション）に追従する責任がクライアントアプリ側に発生します。また、誤ってデータを書き換えないよう、**読み取り専用（readonly）** として設定してください。

## 1. データベース接続設定

`config/database.yml` に認証サーバーのデータベース接続情報を追加します。
ここでは `auth_db` という名前で定義します。

```yaml
default: &default
  adapter: mysql2
  encoding: utf8mb4
  pool: <%= ENV.fetch("RAILS_MAX_THREADS") { 5 } %>
  username: <%= ENV.fetch("DB_USERNAME") { "root" } %>
  password: <%= ENV.fetch("DB_PASSWORD") { "password" } %>
  host: <%= ENV.fetch("DB_HOST") { "localhost" } %>

development:
  primary:
    <<: *default
    database: my_app_development
  # 認証データベースへの接続を追加
  auth_db:
    <<: *default
    database: jyogi_auth # 認証サーバーのDB名
    host: 127.0.0.1      # 認証DBのホスト
    port: 4000           # TiDBのポート (またはMySQLの3306)
    # migrations_paths: db/auth_migrate # マイグレーションを分けたい場合

test:
  primary:
    <<: *default
    database: my_app_test
  auth_db:
    <<: *default
    database: jyogi_auth_test

production:
  primary:
    <<: *default
    database: my_app_production
    password: <%= ENV["MY_APP_DATABASE_PASSWORD"] %>
  auth_db:
    <<: *default
    database: <%= ENV.fetch("AUTH_DB_DATABASE") { "jyogi_auth" } %>
    host: <%= ENV.fetch("AUTH_DB_HOST") { "10.0.0.5" } %> # Cloud SQL Private IP等
    port: 4000
    username: <%= ENV.fetch("AUTH_DB_USERNAME") %>
    password: <%= ENV.fetch("AUTH_DB_PASSWORD") %>
    # 認証DBは読み取り専用レプリカに向けることも検討してください
    # replica: true 
```

## 2. 抽象クラスの作成

認証データベースに接続するための抽象クラスを作成します。
`connects_to` を使用して、このクラス（および継承するモデル）が `auth_db` を使用するように設定します。

`app/models/auth_base.rb`:

```ruby
class AuthBase < ApplicationRecord
  self.abstract_class = true

  # database.yml で定義した auth_db に接続
  connects_to database: { writing: :auth_db, reading: :auth_db }

  # 安全のため読み取り専用にする
  def readonly?
    true
  end
end
```

> [!NOTE]
> `readonly?` を `true` に設定しているため、このモデルを通じて `create`, `update`, `destroy` などを実行しようとすると、`ActiveRecord::ReadOnlyRecord` 例外が発生します。
> これにより、誤って認証データベースのデータを変更してしまう事故を防げます。

## 3. モデルの定義

`AuthBase` を継承して、`User` モデルや `Profile` モデルを定義します。
認証サーバーのテーブル定義に合わせてスキーマを指定します。

`app/models/auth_user.rb`:

```ruby
class AuthUser < AuthBase
  self.table_name = 'users'

  # プロフィールとのリレーション
  has_one :profile, class_name: 'AuthProfile', foreign_key: 'user_id'

  # 主キーがUUIDの場合は型を指定 (Rails 5+)
  attribute :id, :string
  
  # 必要なスコープやメソッドを定義
  scope :active, -> { where.not(last_login_at: nil) }
end
```

`app/models/auth_profile.rb`:

```ruby
class AuthProfile < AuthBase
  self.table_name = 'profiles'

  belongs_to :user, class_name: 'AuthUser', foreign_key: 'user_id'

  attribute :id, :string
end
```

## 4. 利用方法

これで、通常のActive Recordモデルと同じように認証DBのデータを利用できます。

```ruby
# コントローラーなどでの利用例
def index
  # N+1問題を避けるためにincludesを使用
  @users = AuthUser.includes(:profile).all

  @users.each do |user|
    puts "User: #{user.username}"
    puts "Name: #{user.profile&.real_name}"
  end
end
```

## 5. テスト環境での扱い

テスト環境（CIなど）では、Railsアプリ側から認証DBのスキーマをロードできない場合があります（マイグレーションファイルが存在しないため）。

その場合、`db:test:prepare` などのタスクでエラーにならないよう、`database.yml` の `test` 環境設定で `schema_dump: false` を指定するか、モックを使用するなどの工夫が必要になることがあります。

最もシンプルなのは、テスト環境でも認証DB（のダミー）を立ち上げ、認証サーバーのリポジトリからスキーマをインポートすることです。

```bash
# 認証サーバーのリポジトリからスキーマを取得して流し込む例
mysql -h 127.0.0.1 -P 4000 -u root jyogi_auth_test < path/to/jyogi-auth/schema.sql
```
