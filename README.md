# kexecboot.xyz
A Wi-Fi capable network bootloader utilizing a minimal Linux kexec environment and Go TUI, synchronized with netboot.xyz endpoints.

kexecboot.xyz is a bare-metal network bootloader designed to bypass the hardware limitations of traditional iPXE environments. By utilizing a minimal Buildroot Linux initramfs and a statically compiled Go terminal user interface (TUI), kexecboot.xyz enables pre-OS Wi-Fi authentication (WPA2/WPA3) and payload retrieval over modern wireless networks.

The project automatically synchronizes with the upstream endpoints.yml from netboot.xyz, dynamically parsing OS kernel paths and parameters to execute a direct kexec pivot into the target operating system.

Core Capabilities

    Native Wireless Booting: Authenticates to WPA2/WPA3 networks using iwd and modern Linux drivers, bypassing UEFI SNP network constraints.

    Go-Driven Menu System: Replaces legacy iPXE scripts with a statically compiled Go TUI for dynamic OS selection and parameter injection.

    Automated Upstream Sync: CI/CD pipelines automatically fetch, parse, and commit downstream changes from the netboot.xyz operating system registries.

    kexec Execution Handoff: Downloads remote kernels and initial ramdisks directly into memory and pivots CPU execution without rebooting.

Architecture
The deployment artifact is a single minimal Linux .iso or .img designed to be flashed to local media (USB/Disk). Upon boot, the environment initializes the network hardware, fetches the required netboot configurations over HTTP/HTTPS, and hands off system execution to the selected payload.

## Project Structure

```text
.
├── .github/workflows/          # CI/CD pipelines (ISO builds, Upstream sync)
├── buildroot-external/         # Custom Buildroot tree for the bootloader OS
│   ├── board/kexecboot/        # Board-specific configurations
│   │   ├── linux.config        # Minimal Linux kernel config with kexec enabled
│   │   └── rootfs-overlay/     # Custom overlay (e.g., init scripts starting the TUI)
│   ├── configs/                # Buildroot defconfigs (e.g., kexecboot_defconfig)
│   └── package/kexecboot-tui/  # Custom package definition to compile and install our Go TUI
├── tui/                        # The Go-based Terminal User Interface (TUI) source code
│   ├── cmd/kexecboot/          # Application entrypoint
│   └── internal/               
│       ├── menu/               # Bubbletea TUI logic (main menu, prompts)
│       ├── network/            # iwd Wi-Fi authentication integration
│       └── payload/            # netboot.xyz endpoint parsing and kexec execution
└── README.md                   # Project documentation
```
