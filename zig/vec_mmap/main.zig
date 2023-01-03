const std = @import("std");

const allocator = std.heap.page_allocator;

pub const MmapFile = struct {
    ro_bytes: []align(std.mem.page_size) u8,
    len: u64,

    pub fn init(fd: std.os.fd_t) !MmapFile {
        var stat = try std.os.fstat(fd);
        var bytes = try std.os.mmap(null, @intCast(usize, stat.size), std.os.linux.PROT.READ, std.os.MAP.SHARED, fd, 0);
        return MmapFile{ .ro_bytes = bytes, .len = @intCast(u64, stat.size) };
    }

    pub fn deinit(self: *MmapFile) void {
        std.os.munmap(self.ro_bytes);
        self.ro_bytes = undefined;
        self.len = 0;
    }
};

pub fn main() !void {
    var args = try std.process.argsWithAllocator(allocator);
    defer args.deinit();
    _ = try args.next(allocator) orelse @panic("argc == 0");
    const filename = try args.next(allocator) orelse @panic("no file specified");
    std.log.info("Filename: {s}", .{filename});
    var file = try std.fs.cwd().openFile(filename, .{});
    var f = try MmapFile.init(file.handle); // this panics if the file is empty
    defer f.deinit();
    std.log.info("Tamanho: {}", .{f.ro_bytes.len});
}
