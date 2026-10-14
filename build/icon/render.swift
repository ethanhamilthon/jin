// Renders icon-1024.png from the same shapes as icon.svg. Run: swift render.swift
import AppKit

let size = 1024
let rep = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: size, pixelsHigh: size, bitsPerSample: 8,
    samplesPerPixel: 4, hasAlpha: true, isPlanar: false, colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0)!
NSGraphicsContext.saveGraphicsState()
NSGraphicsContext.current = NSGraphicsContext(bitmapImageRep: rep)
NSColor(srgbRed: 0x52 / 255, green: 0xa8 / 255, blue: 0xff / 255, alpha: 1).setFill()
NSBezierPath(roundedRect: NSRect(x: 100, y: 100, width: 824, height: 824), xRadius: 185, yRadius: 185).fill()
let font = NSFont(name: "Times New Roman", size: 515) ?? NSFont.systemFont(ofSize: 515)
let text = NSAttributedString(string: "j", attributes: [.font: font, .foregroundColor: NSColor.black])
let width = text.size().width
text.draw(at: NSPoint(x: 512 - width / 2, y: 1024 - 693 + font.descender))
NSGraphicsContext.restoreGraphicsState()
try! rep.representation(using: .png, properties: [:])!.write(to: URL(fileURLWithPath: "icon-1024.png"))
