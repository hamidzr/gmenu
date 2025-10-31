
  just running gmenu over and over sometimes it gets this nil potiner dref; sometimes it
  hangs with "loading" as a menu item and sometimes exists with proper message

md@wbp ➜ p/gmenu git:(main) ✗  gmenu
ERRO[0000] No items provided through standard input
DEBU[0000] setting exit code to: Exit code 1
INFO[0000] Pid file removed successfully:/var/folders/nq/4ds_61sd2cz60x6vhmsth70c0000gr/T//gmenu.pid
DEBU[0000] Pid file already absent, skipping remove:/var/folders/nq/4ds_61sd2cz60x6vhmsth70c0000gr/T//gmenu.pid
DEBU[0000] Pid file already absent, skipping remove:/var/folders/nq/4ds_61sd2cz60x6vhmsth70c0000gr/T//gmenu.pid
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x2 addr=0x20 pc=0x103411478]

goroutine 1 [running, locked to thread]:
github.com/go-gl/glfw/v3.3/glfw.(*Window).SetCloseCallback(0x14001028068?, 0x140000021c0?)
        /Users/hmd/go/pkg/mod/github.com/go-gl/glfw/v3.3/glfw@v0.0.0-20250301202403-da16c1255728/window.go:781 +0x18
fyne.io/fyne/v2/internal/driver/glfw.(*window).create.func3()
        /Users/hmd/go/pkg/mod/fyne.io/fyne/v2@v2.5.5/internal/driver/glfw/window_desktop.go:771 +0x70
fyne.io/fyne/v2/internal/driver/glfw.(*gLDriver).runGL(0x14000312750)
        /Users/hmd/go/pkg/mod/fyne.io/fyne/v2@v2.5.5/internal/driver/glfw/loop.go:133 +0x178
fyne.io/fyne/v2/internal/driver/glfw.(*gLDriver).Run(0x14000312750)
        /Users/hmd/go/pkg/mod/fyne.io/fyne/v2@v2.5.5/internal/driver/glfw/driver.go:164 +0x78
fyne.io/fyne/v2/app.(*fyneApp).Run(0x1400024ef00)
        /Users/hmd/go/pkg/mod/fyne.io/fyne/v2@v2.5.5/app/app.go:71 +0x6c
github.com/hamidzr/gmenu/core.(*GMenu).RunAppForever(0x140000f6a00)
        /Users/hmd/projects/gmenu/core/gmenurun.go:139 +0xac
github.com/hamidzr/gmenu/internal/cli.run(0x14000252620)
        /Users/hmd/projects/gmenu/internal/cli/cli.go:114 +0x1c8
github.com/hamidzr/gmenu/internal/cli.InitCLI.func1(0x1400030c008, {0x1035e227c?, 0x4?, 0x1035e218e?})
        /Users/hmd/projects/gmenu/internal/cli/cli.go:64 +0x114
github.com/spf13/cobra.(*Command).execute(0x1400030c008, {0x140000340d0, 0x0, 0x0})
        /Users/hmd/go/pkg/mod/github.com/spf13/cobra@v1.10.1/command.go:1015 +0x7d4
github.com/spf13/cobra.(*Command).ExecuteC(0x1400030c008)
        /Users/hmd/go/pkg/mod/github.com/spf13/cobra@v1.10.1/command.go:1148 +0x350
github.com/spf13/cobra.(*Command).Execute(0x1035f039b?)
        /Users/hmd/go/pkg/mod/github.com/spf13/cobra@v1.10.1/command.go:1071 +0x1c
main.main()
        /Users/hmd/projects/gmenu/cmd/main.go:17 +0x40
hmd@wbp ➜ p/gmenu git:(main) ✗  gmenu
ERRO[0000] No items provided through standard input
DEBU[0000] setting exit code to: Exit code 1
INFO[0000] Pid file removed successfully:/var/folders/nq/4ds_61sd2cz60x6vhmsth70c0000gr/T//gmenu.pid
DEBU[0000] Pid file already absent, skipping remove:/var/folders/nq/4ds_61sd2cz60x6vhmsth70c0000gr/T//gmenu.pid
TRAC[0000] Quitting gmenu with code: Exit code 1
hmd@wbp ➜ p/gmenu git:(main) ✗  gmenu
