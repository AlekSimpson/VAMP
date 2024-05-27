// Parts of the layout
// Top Bar
//   - playback controls: skip song, pause/play, repeat, shuffle
//      + makePlaybackControl() -> fyne.CanvasObject
//   - playback view: audio cover, progress bar, name of audio and author
//      + makePlaybackView() -> fyne.CanvasObject
//  + makePlayback() -> fyne.CanvasObject
// Left Side Bar
//   - button to go to sign in/sign out
//      + makeAccountButton() -> fyne.CanvasObject
//   - Lists all audio sets
//      + makeAudioSetsListView() -> fyne.CanvasObject
// Middle Content: audio set
//  + makeAudioListView() -> fyne.CanvasObject

use gdk::Display;
use gtk::prelude::*;
use gtk::*;

const APP_ID: &str = "org.vamp.rust-client";

fn build_scrollview() -> ScrolledWindow {
    // Create a `ListBox` and add labels with integers from 0 to 100
    let list_box = ListBox::new();
    for number in 0..=100 {
        let label = Label::new(Some(&number.to_string()));
        list_box.append(&label);
    }

    let scrolled_window = ScrolledWindow::builder()
        .hscrollbar_policy(gtk::PolicyType::Never) // Disable horizontal scrolling
        .min_content_width(360)
        .child(&list_box)
        .build();

    return scrolled_window;
}

fn build_ui(app: &Application) {
    let container = gtk::Box::builder()
        .css_name("container")
        .build();

    let list_content = build_scrollview();
    let top_label = Label::builder()
        .label("top bar")
        .build();
    let side_label = Label::builder()
        .label("side bar")
        .build();

    let window = ApplicationWindow::builder()
        .application(app)
        .title("VAMP")
        .default_width(800)
        .default_height(600)
        .child(&container)
        .build();

    window.present();
}

fn load_css() {
    // Load the CSS file and add it to the provider
    let provider = CssProvider::new();
    provider.load_from_string(include_str!("styles.css"));

    // Add the provider to the default screen
    gtk::style_context_add_provider_for_display(
        &Display::default().expect("Could not connect to a display."),
        &provider,
        gtk::STYLE_PROVIDER_PRIORITY_APPLICATION,
    );
}

fn main() -> glib::ExitCode {
    // Create a new application
    let app = Application::builder().application_id(APP_ID).build();

    app.connect_startup(|_| load_css());
    app.connect_activate(build_ui);

    // Run the application
    app.run()
}
