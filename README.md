# dsync

Drive-Sync is an unofficial application designed to seamlessly watch and upload changes in files or folders to Google Drive. With Drive-Sync, you can automatically synchronize your local files and folders with your Google Drive account, ensuring that your data is always backed up and up-to-date.

## Features
- **Automatic Synchronization**: Monitors specified directories for changes and uploads new or modified files to Google Drive.
- **Folder and File Tracking**: Currently tracks and syncs creates, deletes, moves, and renames of files and folders.
- **Configurable Watch Directories**: Set up and manage the directories you want to watch for changes.

## Development Status

Please note that Drive-Sync is currently in the development phase. The application is being actively worked on, and while the core functionality for monitoring creates, deletes, moves, and renames is implemented, additional features and improvements are planned for future releases.

## Installation

To install Drive-Sync, follow these steps:

1. **Enable the COPR repository**:

    ```bash
    sudo dnf copr enable khanmf/drive-sync
    ```

2. **Install Drive-Sync**:

    ```bash
    sudo dnf install drive-sync
    ```

## Usage

Once installed, you can use Drive-Sync with the following commands:

1. **Login to Google Drive**:
    ```bash
    dsync login
    ```
    This command will authenticate your Google account and set up the connection.


2. **Add Directories to Watch List for Sync**:

    ```bash
    dsync add dir <Path> <Path>...
    ```
Add one or more directories to the watch list for synchronization.


3. **Get Watched Directories**:

    ```bash
    dsync get list -d
    ```
    Use this command to list all watched directories.


4. **Get Watched Files**:

    ```bash
    dsync get list -f
    ```
    List all watched files.


5. **Get Uploaded Entries**:

    ```bash
    dsync get list -u
    ```
    List all uploaded files and directories.


6. **Get Not Uploaded Entries**:

    ```bash
    
    dsync get list -n
    ```
    List all files and directories that have not been uploaded.


7. **Get Modified Entries**:

    ```bash
    dsync get list -m
    ```
    List all modified files and directories.


8. Get Unmodified Entries:

    ```bash
    dsync get list -s
    ```
    List all unmodified files and directories.

## Important Notes

Upon logging in, Drive-Sync will create a Computer directory and a Computer/{host} directory in your Google Drive. It will then upload the directories on the watch list, maintaining their absolute paths. The same applies to files.

## Contributing

Contributions to Drive-Sync are welcome! If you have suggestions, bug reports, or enhancements, please create an issue or submit a pull request on the Repository.

## License

This project is licensed under the MIT License - see the [LICENCE](./LICENCE.md) file for details.

## Disclaimer

Drive-Sync is an unofficial application and is not affiliated with Google. Use this tool at your own risk. Ensure you comply with Google Drive’s API usage policies and terms of service.
