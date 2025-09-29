const std = @import("std");

pub fn main() !void {
    var allocator = std.heap.GeneralPurposeAllocator(.{}){};
    defer std.debug.assert(allocator.deinit() == .ok);
    var httpPort = try std.fmt.parseInt(i32, std.os.getenv("PORT") orelse "3001", 10);
    std.debug.print("Listening on port: {}\n", .{httpPort});
    // TODO: how to use the client in zig
}
