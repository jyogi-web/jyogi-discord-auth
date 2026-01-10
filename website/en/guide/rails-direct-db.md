# Rails: Direct Database Access (Multiple Databases)

This guide explains how to directly access the Jyogi Auth System database (e.g., `users` table) using the [Multiple Databases](https://guides.rubyonrails.org/active_record_multiple_databases.html) feature introduced in Ruby on Rails 6.0.

> [!WARNING]
> **Recommendation**: We strongly recommend using [OAuth2 API Integration](./client-integration.md) instead.
> If you access the database directly, the client app becomes responsible for following schema changes (migrations) in the Auth Server. Also, ensure to configure it as **readonly** to prevent accidental data modification.

## 1. Database Configuration

Add the connection information for the Auth Server's database to `config/database.yml`.
Here, we define it as `auth_db`.

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
  # Add connection to Auth Database
  auth_db:
    <<: *default
    database: jyogi_auth # Auth Server DB name
    host: 127.0.0.1      # Auth DB Host
    port: 4000           # TiDB Port (or 3306 for MySQL)
    # migrations_paths: db/auth_migrate # If you want separate migrations

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
    host: <%= ENV.fetch("AUTH_DB_HOST") { "10.0.0.5" } %> # e.g., Cloud SQL Private IP
    port: 4000
    username: <%= ENV.fetch("AUTH_DB_USERNAME") %>
    password: <%= ENV.fetch("AUTH_DB_PASSWORD") %>
    # Consider pointing to a read-only replica
    # replica: true 
```

## 2. Create Abstract Class

Create an abstract class to connect to the Auth Database.
Use `connects_to` to configure this class (and models inheriting from it) to use `auth_db`.

`app/models/auth_base.rb`:

```ruby
class AuthBase < ApplicationRecord
  self.abstract_class = true

  # Connect to auth_db defined in database.yml
  connects_to database: { writing: :auth_db, reading: :auth_db }

  # Set to readonly for safety
  def readonly?
    true
  end
end
```

> [!NOTE]
> Since `readonly?` is set to `true`, attempting to execute `create`, `update`, or `destroy` through this model will raise an `ActiveRecord::ReadOnlyRecord` exception.
> This prevents accidental modification of the data in the Auth Database.

## 3. Define Models

Define `User` and `Profile` models by inheriting from `AuthBase`.
Specify the schema according to the Auth Server's table definitions.

`app/models/auth_user.rb`:

```ruby
class AuthUser < AuthBase
  self.table_name = 'users'

  # Relation with Profile
  has_one :profile, class_name: 'AuthProfile', foreign_key: 'user_id'

  # Specify type if PK is UUID (Rails 5+)
  attribute :id, :string
  
  # Define necessary scopes and methods
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

## 4. Usage

Now you can use the data from the Auth DB just like regular Active Record models.

```ruby
# Example usage in a controller
def index
  # Use includes to avoid N+1 problem
  @users = AuthUser.includes(:profile).all

  @users.each do |user|
    puts "User: #{user.username}"
    puts "Name: #{user.profile&.real_name}"
  end
end
```

## 5. Handling in Test Environment

In test environments (like CI), the Rails app might fail to load the Auth DB schema because the migration files do not exist in the Rails app.

To avoid errors with tasks like `db:test:prepare`, you might need to set `schema_dump: false` in the `test` environment configuration of `database.yml`, or use mocks.

The simplest approach is to launch a dummy Auth DB in the test environment and import the schema from the Auth Server repository.

```bash
# Example: Import schema from Auth Server repo
mysql -h 127.0.0.1 -P 4000 -u root jyogi_auth_test < path/to/jyogi-auth/schema.sql
```
