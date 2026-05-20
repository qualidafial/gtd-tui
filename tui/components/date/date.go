package date

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/sho0pi/naturaltime"
)

// Field is a huh.Field that edits an optional date/time bound to a *time.Time.
// Empty input means nil. Accepted formats:
//   - YYYY-MM-DD
//   - YYYY-MM-DD HH:MM
//   - YYYY-MM-DD HH:MM:SS
//   - natural-language expressions parsed by naturaltime
//     (e.g. "tomorrow", "in 3 hours", "next monday at 9am")
//
// All values are interpreted in the local time zone.
type Field struct {
	accessor huh.Accessor[*time.Time]
	key      string

	title       string
	description string
	placeholder string

	textinput textinput.Model

	validate func(*time.Time) error
	err      error
	focused  bool

	width  int
	height int

	theme     huh.Theme
	hasDarkBg bool
	keymap    huh.InputKeyMap
}

// NewField creates a new date field.
func NewField() *Field {
	ti := textinput.New()
	ti.Placeholder = "YYYY-MM-DD [HH:MM[:SS]] or e.g. \"tomorrow\""

	return &Field{
		accessor:    &huh.EmbeddedAccessor[*time.Time]{},
		textinput:   ti,
		validate:    func(*time.Time) error { return nil },
		placeholder: ti.Placeholder,
	}
}

func (d *Field) Value(value **time.Time) *Field {
	return d.Accessor(huh.NewPointerAccessor(value))
}

func (d *Field) Accessor(a huh.Accessor[*time.Time]) *Field {
	d.accessor = a
	d.textinput.SetValue(formatDate(d.accessor.Get()))
	return d
}

func (d *Field) Key(k string) *Field         { d.key = k; return d }
func (d *Field) Title(s string) *Field       { d.title = s; return d }
func (d *Field) Description(s string) *Field { d.description = s; return d }
func (d *Field) Placeholder(s string) *Field {
	d.placeholder = s
	d.textinput.Placeholder = s
	return d
}
func (d *Field) Validate(fn func(*time.Time) error) *Field {
	d.validate = fn
	return d
}

// huh.Field interface ------------------------------------------------------

func (d *Field) Error() error { return d.err }
func (*Field) Skip() bool     { return false }
func (*Field) Zoom() bool     { return false }

func (d *Field) Focus() tea.Cmd {
	d.focused = true
	return d.textinput.Focus()
}

func (d *Field) Blur() tea.Cmd {
	d.focused = false
	d.textinput.Blur()
	t, err := parseDate(d.textinput.Value())
	if err != nil {
		d.err = err
		return nil
	}
	d.accessor.Set(t)
	d.textinput.SetValue(formatDate(t))
	d.err = d.validate(t)
	return nil
}

func (d *Field) KeyBinds() []key.Binding {
	return []key.Binding{d.keymap.Prev, d.keymap.Submit, d.keymap.Next}
}

func (d *Field) Init() tea.Cmd {
	d.textinput.Blur()
	return nil
}

func (d *Field) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		d.hasDarkBg = msg.IsDark()
	case tea.KeyPressMsg:
		d.err = nil
		switch {
		case key.Matches(msg, d.keymap.Prev):
			cmds = append(cmds, huh.PrevField)
		case key.Matches(msg, d.keymap.Next, d.keymap.Submit):
			t, err := parseDate(d.textinput.Value())
			if err != nil {
				d.err = err
				return d, nil
			}
			if err := d.validate(t); err != nil {
				d.err = err
				return d, nil
			}
			d.accessor.Set(t)
			d.textinput.SetValue(formatDate(t))
			cmds = append(cmds, huh.NextField)
		}
	}

	var cmd tea.Cmd
	d.textinput, cmd = d.textinput.Update(msg)
	cmds = append(cmds, cmd)
	return d, tea.Batch(cmds...)
}

func (d *Field) activeStyles() *huh.FieldStyles {
	theme := d.theme
	if theme == nil {
		theme = huh.ThemeFunc(huh.ThemeCharm)
	}
	if d.focused {
		return &theme.Theme(d.hasDarkBg).Focused
	}
	return &theme.Theme(d.hasDarkBg).Blurred
}

func (d *Field) View() string {
	styles := d.activeStyles()
	maxWidth := d.width - styles.Base.GetHorizontalFrameSize()

	st := d.textinput.Styles()
	st.Cursor.Color = styles.TextInput.Cursor.GetForeground()
	st.Focused.Prompt = styles.TextInput.Prompt
	st.Focused.Text = styles.TextInput.Text
	st.Focused.Placeholder = styles.TextInput.Placeholder
	d.textinput.SetStyles(st)

	var sb strings.Builder
	if d.title != "" {
		sb.WriteString(styles.Title.Render(d.title))
		sb.WriteString("\n")
	}
	if d.description != "" {
		sb.WriteString(styles.Description.Render(d.description))
		sb.WriteString("\n")
	}
	sb.WriteString(d.textinput.View())

	_ = maxWidth // reserved for wrapping if title/description grow long
	return styles.Base.Width(d.width).Height(d.height).Render(sb.String())
}

func (d *Field) Run() error { return huh.Run(d) }

func (d *Field) RunAccessible(w io.Writer, r io.Reader) error {
	// Minimal accessible implementation: print prompt, read one line, parse.
	prompt := d.activeStyles().Title.PaddingRight(1).Render(d.title)
	if prompt == "" {
		prompt = "Date: "
	}
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return err
	}
	var line string
	if _, err := fmt.Fscanln(r, &line); err != nil && err != io.EOF {
		return err
	}
	t, err := parseDate(line)
	if err != nil {
		return err
	}
	d.accessor.Set(t)
	return nil
}

func (d *Field) WithKeyMap(k *huh.KeyMap) huh.Field {
	d.keymap = k.Input
	return d
}

func (d *Field) WithTheme(theme huh.Theme) huh.Field {
	if d.theme != nil {
		return d
	}
	d.theme = theme
	return d
}

func (d *Field) WithWidth(width int) huh.Field {
	styles := d.activeStyles()
	d.width = width
	frame := styles.Base.GetHorizontalFrameSize()
	promptWidth := lipgloss.Width(d.textinput.Styles().Focused.Prompt.Render(d.textinput.Prompt))
	d.textinput.SetWidth(width - frame - promptWidth - 1)
	return d
}

func (d *Field) WithHeight(height int) huh.Field {
	d.height = height
	return d
}

func (d *Field) WithPosition(p huh.FieldPosition) huh.Field {
	d.keymap.Prev.SetEnabled(!p.IsFirst())
	d.keymap.Next.SetEnabled(!p.IsLast())
	d.keymap.Submit.SetEnabled(p.IsLast())
	return d
}

func (d *Field) GetKey() string { return d.key }
func (d *Field) GetValue() any  { return d.accessor.Get() }

// Parsing / formatting -----------------------------------------------------

var dateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

// naturalParser is the lazily-initialized naturaltime parser. Initialization
// compiles an embedded JS program in a goja runtime, so we do it once.
var naturalParser = sync.OnceValues(naturaltime.New)

// parseDate returns nil for empty input, or a parsed local time.
func parseDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}

	parser, err := naturalParser()
	if err != nil {
		return nil, fmt.Errorf("init natural-language date parser: %w", err)
	}
	t, err := parser.ParseDate(s, time.Now())
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", s, err)
	}
	if t == nil {
		return nil, fmt.Errorf("invalid date %q", s)
	}
	return t, nil
}

// formatDate renders nil as "" and chooses date-only vs date+time based on
// whether the time component is midnight local.
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	local := t.Local()
	if local.Hour() == 0 && local.Minute() == 0 && local.Second() == 0 {
		return local.Format("2006-01-02")
	}
	return local.Format("2006-01-02 15:04")
}
