package cogent

import (
	"cogentcore.org/core/core"
)

// RunExample launches a centered text input demo using Cogent Core.
func RunExample() {
	body := core.NewBody("Centered Text Input")
	body.SetTitle("Centered Text Input")

	root := core.NewFrame(body)
	// root.Styler(func(s *styles.Style) {
	// 	s.Direction = styles.Column
	// 	s.CenterAll()
	// 	s.Grow.SetScalar(1)
	// 	s.Min.Set(units.Pw(100), units.Ph(100))
	// 	s.Gap.Set(units.Dp(12))
	// })
	//
	core.NewText(root).SetText("Type something:")
	// tf := core.NewTextField(root)
	// tf.SetPlaceholder("Hello, Cogent Core!")

	body.RunMainWindow()
}

/*
Optional: control OS-level window geometry instead of relying on saved defaults.

	st := body.NewWindow()
	st.SetTitle("Centered Text Input")
	st.SetSize(units.Dp(520), units.Dp(280))
	// st.SetPos(image.Pt(x, y))
	st.RunMain()
*/
