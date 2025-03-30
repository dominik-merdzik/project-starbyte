<h1 align="center">💫 Project Starbyte 💫 </h1>

> The year is 2399 and humanity has reached the stars. Space travel beyond the Solar System is now possible with the use of Faster-Than-Light (FTL) technology. They are not alone in the universe. Contact has been made with intelligent alien species. Some are friendly, some hostile, and some are indifferent. Humanity is now part of a galactic community, but the galaxy is a mysterious place. Your ship and crew are ready, and the stars await…

Project Starbyte is a terminal game where the player commands a space vessel with Faster-Than-Light capabilities. Explore the galaxy with your crew, experience story-driven missions, and discover the mysteries of the cosmos.

Built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea/) TUI framework.

## Running the game

Users can play Starbyte in two ways. Option #1 is recommended for most users because it’s easier. Option #2 is available if users want to modify the source code.

The game has very minimal system requirements because the game lacks graphics. Essentially, any modern computer can run this game with excellent performance. These systems are supported:

| Operating System              | CPU Architecture              |
| ----------------------------- | ----------------------------- |
| Microsoft Windows 10 or later | x86-64, ARM64                 |
| Apple macOS 11 or later       | x86-64, ARM64 (Apple Silicon) |
| Linux 2.6.32 or later         | x86-64, ARM64                 |

The game should also function on any terminal environment, with minor appearance differences. These terminals are confirmed to work:

- Windows Terminal
- PowerShell
- Command Prompt
- Zsh

### Option 1: Download and run the binary (recommended)

1. Go to the Releases page on the GitHub repository.
2. Download the latest version for your operating system and CPU architecture.
3. Run the executable.

Note: macOS users will be blocked from running the executable at first. Go to the “Privacy and Security” settings and allow the app to run.

### Option 2: Build from source

This approach requires the user to have Go v1.23 installed.

1. Git pull the repository https://github.com/dominik-merdzik/project-starbyte
2. Open at terminal at `cmd/project-starbyte/` and run `go run .`

## Configuration

User configuration through the main menu is coming soon! For now, you can edit music settings (including disabling) in `cmd/project-starbyte/config/config.toml`.
