const std = @import("std");

const allocator = std.heap.page_allocator;

pub fn main() !void {
    var timestamp = std.time.milliTimestamp();
    const str = "eoq trabson";
    std.log.info("Tamanho: {}", .{str.len});
    var valAmt = @bitCast(usize, @rem(timestamp, @as(i64, 256)));
    const val = try allocator.alloc(u8, valAmt);
    std.log.info("Tamanho: {}", .{val.len});
}
