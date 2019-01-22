# Terraform Enterprise

Atlantis integrates seamlessly with Terraform Enterprise's new [Free Remote State Management](https://app.terraform.io/signup).

[[toc]]

## Migrating to TFE's Remote State
If you're using a different state backend, you first need to migrate your state
to use Terraform Enterprise (TFE). Read [TODO: use right link](https://www.terraform.io/docs/enterprise/migrate/index.html)
for more information on how to migrate.

## Configuring Atlantis
Once you've migrated your state to TFE, and your code is using the TFE backend,
you're ready to configure Atlantis.

Atlantis needs a TFE Token that it will use to access the TFE API.
Using a **Team Token is recommended**, however you can also use a User Token.

### Team Token
To generate a team token, click on **Settings** in the top bar, then **Teams** in
the sidebar, then scroll down to **Team API Token**.

### User Token
To generate a user token, click on your avatar, then **User Settings**, then
**Tokens** in the sidebar.

### Passing The Token To Atlantis
The token can be passed to Atlantis via the `ATLANTIS_TFE_TOKEN` environment variable.

You can also use the `--tfe-token` flag, however your token would then be easily
viewable in the process list.

That's it! Atlantis should be able to perform Terraform operations using TFE's
remote state backend now.

:::tip NOTE
Under the hood, Atlantis is generating a `~/.terraformrc` file.
If you already had a `~/.terraformrc` file, then you'll need to manually
add the credentials block:
```
...
credentials "app.terraform.io" {
  token = "xxxx"
}
```
Instead of using the `ATLANTIS_TFE_TOKEN` environment variable since Atlantis
won't overwrite your `.terraformrc` file.
:::
