from PyQt6.QtWidgets import QPlainTextEdit, QMainWindow, QApplication, QMessageBox, QMenuBar, QMenu
from PyQt6.QtGui import QKeySequence, QPalette, QColor, QAction
from PyQt6.QtCore import Qt


class MainWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        text = QPlainTextEdit()
        self.setCentralWidget(text)

        self.setMenuBar(self._get_menu_bar())

    def _get_menu_bar(self):
        bar = QMenuBar(self)

        help_menu = bar.addMenu("Help")

        action_help_about = QAction("About", self)
        action_help_about.triggered.connect(self.about_dialog)
        help_menu.addAction(action_help_about)
        return bar


    def about_dialog(self):
        text = "<center>" \
               "<h1>Text Editor</h1>" \
               "&#8291;" \
               "<img src=icon.svg>" \
               "</center>" \
               "<p>Version 31.4.159.265358<br/>" \
               "Copyright &copy; Company Inc.</p>"
        QMessageBox.about(self, "About Text Editor", text)

def main():
    app = QApplication([])
    app.setStyle('Fusion')
    app.setApplicationName("wawawawa")

    window = MainWindow()

    window.show()
    app.exec()

if __name__ == "__main__":
    main()
