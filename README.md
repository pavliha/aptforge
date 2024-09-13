# AptForge

**AptForge** is an open-source command-line tool for managing custom APT repositories, designed to streamline the upload of `.deb` packages and automate the generation of repository metadata files such as `Packages` and `Release`.

## Features

- **Upload `.deb` Files**: Upload Debian packages to S3-compatible storage (e.g., AWS S3, DigitalOcean Spaces).
- **Metadata Management**: Automatically update `Packages` and `Release` files with correct checksums.
- **Environment Variable Support**: Use environment variables for access credentials.
- **Customizable Repository Configurations**: Set custom repository component, origin, label, architecture, and archive type.

## Installation

Clone the repository and build the binary using Go:

```bash
git clone https://github.com/your-username/aptforge.git
cd aptforge
go build -o aptforge ./cmd
```

Alternatively, you can download a pre-built binary from the releases page.

## Usage
The basic usage involves uploading a .deb package to an S3-compatible storage and updating the repository metadata.

```bash
aptforge --file /path/to/package.deb --bucket my-bucket \
--access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY \
--endpoint fra1.digitaloceanspaces.com
Flags
Flag	Description	Required	Default
--file	Path to the .deb file to upload	Yes
--bucket	Name of the S3 bucket	Yes
--access-key	Access Key for the S3 bucket	Yes
--secret-key	Secret Key for the S3 bucket	Yes
--endpoint	S3-compatible endpoint (e.g., fra1.digitaloceanspaces.com)	Yes
--component	Repository component (e.g., main, contrib, non-free)	No	main
--origin	Origin of the repository	No	Custom Repository
--label	Label for the repository	No	Custom Repo
--arch	Target architecture for the repository (e.g., amd64, arm64)	No	amd64
--archive	Archive type of the repository (e.g., stable, testing, unstable)	No	stable
```
Example
```bash
aptforge --file ./my-package.deb --bucket my-repo-bucket \
--access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY \
--endpoint fra1.digitaloceanspaces.com --component main \
--origin "My Custom Repo" --label "My Repo Label" --arch amd64
```
## Environment Variables
AptForge can use environment variables for credentials. If --access-key or --secret-key are not provided via flags, the tool will look for:

`AWS_ACCESS_KEY_ID`
`AWS_SECRET_ACCESS_KEY`

## Error Handling
AptForge will log detailed error messages using logrus if it encounters any issues during the execution, such as failure to upload files, missing credentials, or invalid paths.

## Contributing
We welcome contributions! Follow these steps to contribute to AptForge:

## Fork the repository.
Create a feature branch: git checkout -b feature/my-feature.
Commit your changes: git commit -m 'Add some feature'.
Push to the branch: git push origin feature/my-feature.
Open a pull request.
Please ensure that your code follows our coding guidelines and includes appropriate tests.

## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Issues
If you encounter any issues or bugs, feel free to open a GitHub issue here. Please provide a detailed description of the problem along with steps to reproduce it.

## Roadmap
Add support for GPG signing of Release files.
Implement automatic retries for S3 upload failures.
Extend support for other architectures (e.g., arm64).
Stay tuned for updates!

## Maintainers
Pavlo Kostiuk (@pavliha)
