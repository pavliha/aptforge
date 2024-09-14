# AptForge

**AptForge** is an open-source command-line tool for managing custom APT repositories. It streamlines the upload of .deb packages to S3-compatible storage and automates the generation of repository metadata files such as Packages, Packages.gz, and Release.

## Features

- **Upload `.deb` Files**: Seamlessly upload Debian packages to S3-compatible storage (e.g., AWS S3, DigitalOcean Spaces, MinIO).
- **Metadata Management**: Automatically update Packages, Packages.gz, and Release files with correct checksums.
- **Environment Variable Support**: Use environment variables for access credentials if flags are not provided.
- **Customizable Repository Configurations**: Set custom repository component, origin, label, architecture, and archive type.
- **Secure Connections**: Enable or disable secure connections based on your storage endpoint requirements.

## Installation

### Using `go install`

Clone the repository and build the binary using Go:
You can install AptForge directly using the go install command:

```bash
go install github.com/pavliha/aptforge@latest
```
This command will download and install the latest version of AptForge. Ensure that your $GOPATH/bin (or $HOME/go/bin if $GOPATH is not set) is added to your system's PATH environment variable so that you can run the aptforge command from anywhere.

### Building from Source
Clone the repository and build the binary using Go:

```bash
git clone https://github.com/pavliha/aptforge.git
cd aptforge
go build -o aptforge .
```

### Download Pre-built Binary
Alternatively, you can download a pre-built binary from the [releases](https://github.com/pavliha/aptforge/releases) page.


## Usage
The basic usage involves uploading a .deb package to an S3-compatible storage and updating the repository metadata.

```bash
aptforge --file /path/to/package.deb --bucket my-bucket \
--access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY \
--endpoint fra1.digitaloceanspaces.com
```


### Example
```bash
aptforge --file ./my-package.deb --bucket my-repo-bucket \
--access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY \
--endpoint fra1.digitaloceanspaces.com --component main \
--origin "My Custom Repo" --label "My Repo Label" --arch amd64
```

### Example with All Options

```bash
aptforge --file ./my-package.deb --bucket my-repo-bucket \
--access-key YOUR_ACCESS_KEY --secret-key YOUR_SECRET_KEY \
--endpoint your-s3-endpoint.com --component main \
--origin "My Custom Repo" --label "My Repo Label" \
--arch amd64 --archive stable --secure=true
```

## Flags
| Flag           | Description                                                            | Required | Default          |
|----------------|------------------------------------------------------------------------|----------|------------------|
| `--file`       | Path to the `.deb` file to upload                                      | Yes      |                  |
| `--bucket`     | Name of the S3 bucket                                                  | Yes      |                  |
| `--access-key` | Access Key for the S3 bucket                                           | Yes      |                  |
| `--secret-key` | Secret Key for the S3 bucket                                           | Yes      |                  |
| `--endpoint`   | S3-compatible endpoint (e.g., `fra1.digitaloceanspaces.com`)           | Yes      |                  |
| `--component`  | Repository component (e.g., `main`, `contrib`, `non-free`)             | No       | `main`           |
| `--origin`     | Origin of the repository                                               | No       | `Apt Repository` |
| `--label`      | Label for the repository                                               | No       | `Apt Repo`       |
| `--arch`       | Target architecture for the repository (e.g., `amd64`, `arm64`)        | No       | `amd64`          |
| `--archive`    | Archive type of the repository (e.g., `stable`, `testing`, `unstable`) | No       | `stable`         |
| `--secure`     | Enable secure connections (true or false)                              | No       | `true`           |

**Note:** If --access-key or --secret-key are not provided via flags, AptForge will look for the environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

### Valid Values
- **Architecture** (--arch): amd64, arm64, i386
- **Archives** (--archive): stable, testing, unstable
- **Components** (--component): main, contrib, non-free

## Environment Variables
AptForge can use environment variables for credentials. If --access-key or --secret-key are not provided via flags, the tool will look for:

`AWS_ACCESS_KEY_ID`
`AWS_SECRET_ACCESS_KEY`

## Error Handling
AptForge uses Logrus for structured logging. Detailed error messages will be logged if the tool encounters any issues during execution, such as failure to upload files, missing credentials, or invalid paths.

## Contributing
All contributions are welcome! Follow these steps to contribute to AptForge:

1. Fork the repository.
2. Create a feature branch: git checkout -b feature/my-feature.
3. Commit your changes: git commit -m 'Add some feature.'
4. Push to the branch: git push origin feature/my-feature.
5. Open a pull request. 

Please ensure that your code follows coding guidelines and includes appropriate tests.

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
