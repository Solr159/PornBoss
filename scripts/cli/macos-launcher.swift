import Cocoa
import Darwin
import Foundation

private let startupURLPattern = try! NSRegularExpression(
  pattern: #"https?://localhost:\d+"#,
  options: []
)

final class OutputBuffer {
  private let limit: Int
  private let lock = NSLock()
  private var data = Data()

  init(limit: Int) {
    self.limit = max(limit, 1024)
  }

  func append(_ chunk: Data) {
    guard !chunk.isEmpty else {
      return
    }
    lock.lock()
    defer { lock.unlock() }
    data.append(chunk)
    if data.count > limit {
      data = Data(data.suffix(limit))
    }
  }

  func stringValue() -> String {
    lock.lock()
    defer { lock.unlock() }
    if let text = String(data: data, encoding: .utf8) {
      return text.trimmingCharacters(in: .whitespacesAndNewlines)
    }
    return String(decoding: data, as: UTF8.self).trimmingCharacters(in: .whitespacesAndNewlines)
  }
}

final class LineBuffer {
  private let lock = NSLock()
  private var pending = ""

  func append(_ chunk: Data) -> [String] {
    guard !chunk.isEmpty else {
      return []
    }
    let text: String
    if let utf8 = String(data: chunk, encoding: .utf8) {
      text = utf8
    } else {
      text = String(decoding: chunk, as: UTF8.self)
    }

    lock.lock()
    defer { lock.unlock() }

    pending += text
    let normalized = pending.replacingOccurrences(of: "\r\n", with: "\n")
    let parts = normalized.components(separatedBy: "\n")
    pending = parts.last ?? ""
    return Array(parts.dropLast()).map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
  }
}

final class AppDelegate: NSObject, NSApplicationDelegate {
  private let outputBuffer = OutputBuffer(limit: 16 * 1024)
  private let stdoutLineBuffer = LineBuffer()
  private let stderrLineBuffer = LineBuffer()
  private var serverProcess: Process?
  private var outputPipes: [Pipe] = []
  private var waitingForTerminationReply = false
  private var serverURL: URL?
  private var hasAutoOpenedPage = false
  private var window: NSWindow?
  private var statusLabel: NSTextField?
  private var openButton: NSButton?

  func applicationDidFinishLaunching(_ notification: Notification) {
    installMenu()
    buildWindow()
    launchServer()
  }

  func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
    true
  }

  func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
    guard let process = serverProcess, process.isRunning else {
      return .terminateNow
    }
    waitingForTerminationReply = true
    terminateServer(process)
    return .terminateLater
  }

  @objc
  private func quitSelected(_ sender: Any?) {
    NSApp.terminate(sender)
  }

  @objc
  private func openPageSelected(_ sender: Any?) {
    guard let url = serverURL else {
      NSSound.beep()
      return
    }
    NSWorkspace.shared.open(url)
  }

  private func installMenu() {
    let appName =
      (Bundle.main.object(forInfoDictionaryKey: "CFBundleName") as? String)?.trimmingCharacters(in: .whitespacesAndNewlines)
      ?? "Pornboss"

    let mainMenu = NSMenu()
    let appMenuItem = NSMenuItem()
    mainMenu.addItem(appMenuItem)

    let appMenu = NSMenu()
    let quitItem = NSMenuItem(title: "Quit \(appName)", action: #selector(quitSelected(_:)), keyEquivalent: "q")
    quitItem.target = self
    appMenu.addItem(quitItem)

    appMenuItem.submenu = appMenu
    NSApp.mainMenu = mainMenu
  }

  private func buildWindow() {
    let window = NSWindow(
      contentRect: NSRect(x: 0, y: 0, width: 360, height: 140),
      styleMask: [.titled, .closable, .miniaturizable],
      backing: .buffered,
      defer: false
    )
    window.center()
    window.title = "Pornboss"
    window.isReleasedWhenClosed = false

    let contentView = NSView(frame: window.contentRect(forFrameRect: window.frame))
    contentView.translatesAutoresizingMaskIntoConstraints = false
    window.contentView = contentView

    let statusLabel = NSTextField(labelWithString: "Starting Pornboss...")
    statusLabel.font = .systemFont(ofSize: 14)
    statusLabel.alignment = .center
    statusLabel.translatesAutoresizingMaskIntoConstraints = false

    let openButton = NSButton(title: "Open Page", target: self, action: #selector(openPageSelected(_:)))
    openButton.bezelStyle = .rounded
    openButton.isEnabled = false
    openButton.translatesAutoresizingMaskIntoConstraints = false

    let quitButton = NSButton(title: "Quit", target: self, action: #selector(quitSelected(_:)))
    quitButton.bezelStyle = .rounded
    quitButton.translatesAutoresizingMaskIntoConstraints = false

    contentView.addSubview(statusLabel)
    contentView.addSubview(openButton)
    contentView.addSubview(quitButton)

    NSLayoutConstraint.activate([
      statusLabel.topAnchor.constraint(equalTo: contentView.topAnchor, constant: 28),
      statusLabel.leadingAnchor.constraint(equalTo: contentView.leadingAnchor, constant: 20),
      statusLabel.trailingAnchor.constraint(equalTo: contentView.trailingAnchor, constant: -20),

      openButton.topAnchor.constraint(equalTo: statusLabel.bottomAnchor, constant: 24),
      openButton.centerXAnchor.constraint(equalTo: contentView.centerXAnchor, constant: -62),
      openButton.widthAnchor.constraint(equalToConstant: 110),

      quitButton.topAnchor.constraint(equalTo: statusLabel.bottomAnchor, constant: 24),
      quitButton.centerXAnchor.constraint(equalTo: contentView.centerXAnchor, constant: 62),
      quitButton.widthAnchor.constraint(equalToConstant: 110),
    ])

    self.window = window
    self.statusLabel = statusLabel
    self.openButton = openButton

    window.makeKeyAndOrderFront(nil)
    NSApp.activate(ignoringOtherApps: true)
  }

  private func updateStatus(_ text: String) {
    statusLabel?.stringValue = text
  }

  private func handleServerLine(_ line: String) {
    guard !line.isEmpty else {
      return
    }
    guard serverURL == nil else {
      return
    }
    let range = NSRange(line.startIndex..<line.endIndex, in: line)
    guard let match = startupURLPattern.firstMatch(in: line, options: [], range: range),
          let matchRange = Range(match.range, in: line),
          let url = URL(string: String(line[matchRange]))
    else {
      return
    }

    serverURL = url
    updateStatus("Pornboss is running.")
    openButton?.isEnabled = true
    if !hasAutoOpenedPage {
      hasAutoOpenedPage = true
      NSWorkspace.shared.open(url)
    }
  }

  private func launchServer() {
    guard let launcherURL = Bundle.main.executableURL else {
      showFatal(message: "Pornboss could not determine its executable path.", detail: "")
      return
    }

    let executableDir = launcherURL.deletingLastPathComponent()
    let serverURL = executableDir.appendingPathComponent("pornboss-server")
    guard FileManager.default.isExecutableFile(atPath: serverURL.path) else {
      showFatal(
        message: "Pornboss is missing the bundled server executable.",
        detail: serverURL.path
      )
      return
    }

    let process = Process()
    process.executableURL = serverURL
    process.currentDirectoryURL = executableDir
    process.arguments = Array(CommandLine.arguments.dropFirst())

    var environment = ProcessInfo.processInfo.environment
    environment["PORNBOSS_LAUNCHER"] = "1"
    process.environment = environment

    let stdoutPipe = Pipe()
    let stderrPipe = Pipe()
    stdoutPipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
      guard let self else {
        return
      }
      let data = handle.availableData
      self.outputBuffer.append(data)
      let lines = self.stdoutLineBuffer.append(data)
      guard !lines.isEmpty else {
        return
      }
      DispatchQueue.main.async {
        for line in lines {
          self.handleServerLine(line)
        }
      }
    }
    stderrPipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
      guard let self else {
        return
      }
      let data = handle.availableData
      self.outputBuffer.append(data)
      let lines = self.stderrLineBuffer.append(data)
      guard !lines.isEmpty else {
        return
      }
      DispatchQueue.main.async {
        for line in lines {
          self.handleServerLine(line)
        }
      }
    }
    process.standardOutput = stdoutPipe
    process.standardError = stderrPipe
    outputPipes = [stdoutPipe, stderrPipe]

    process.terminationHandler = { [weak self] proc in
      DispatchQueue.main.async {
        self?.handleServerExit(proc)
      }
    }

    do {
      try process.run()
      serverProcess = process
      updateStatus("Starting Pornboss...")
    } catch {
      let detail = [serverURL.path, error.localizedDescription]
        .filter { !$0.isEmpty }
        .joined(separator: "\n")
      showFatal(message: "Pornboss failed to start.", detail: detail)
    }
  }

  private func terminateServer(_ process: Process) {
    guard process.isRunning else {
      return
    }
    process.terminate()
    DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
      guard process.isRunning else {
        return
      }
      kill(process.processIdentifier, SIGKILL)
    }
  }

  private func handleServerExit(_ process: Process) {
    for pipe in outputPipes {
      pipe.fileHandleForReading.readabilityHandler = nil
    }
    outputPipes.removeAll()
    serverProcess = nil

    if waitingForTerminationReply {
      NSApp.reply(toApplicationShouldTerminate: true)
      return
    }

    if process.terminationReason == .exit && process.terminationStatus == 0 {
      NSApp.terminate(nil)
      return
    }

    updateStatus("Pornboss stopped.")

    var lines = ["The bundled server stopped unexpectedly."]
    if process.terminationReason == .uncaughtSignal {
      lines.append("Signal: \(process.terminationStatus)")
    } else {
      lines.append("Exit status: \(process.terminationStatus)")
    }

    let capturedOutput = outputBuffer.stringValue()
    if !capturedOutput.isEmpty {
      lines.append("")
      lines.append(capturedOutput)
    }

    showFatal(message: "Pornboss stopped.", detail: lines.joined(separator: "\n"))
  }

  private func showFatal(message: String, detail: String) {
    let alert = NSAlert()
    alert.alertStyle = .critical
    alert.messageText = message
    alert.informativeText = detail
    alert.runModal()
    NSApp.terminate(nil)
  }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.setActivationPolicy(.regular)
app.run()
