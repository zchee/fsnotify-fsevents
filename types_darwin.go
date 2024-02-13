//go:build ignore

package fsevents

/*
#cgo LDFLAGS: -framework CoreServices
#include <CoreServices/CoreServices.h>

static CFArrayRef ArrayCreateMutable(int len) {
	return CFArrayCreateMutable(NULL, len, &kCFTypeArrayCallBacks);
}

extern void fsevtCallback(FSEventStreamRef p0, uintptr_t info, size_t p1, char** p2, FSEventStreamEventFlags* p3, FSEventStreamEventId* p4);

static FSEventStreamRef EventStreamCreateRelativeToDevice(FSEventStreamContext * context, uintptr_t info, dev_t dev, CFArrayRef paths, FSEventStreamEventId since, CFTimeInterval latency, FSEventStreamCreateFlags flags) {
	context->info = (void*) info;
	return FSEventStreamCreateRelativeToDevice(NULL, (FSEventStreamCallback) fsevtCallback, context, dev, paths, since, latency, flags);
}

static FSEventStreamRef EventStreamCreate(FSEventStreamContext * context, uintptr_t info, CFArrayRef paths, FSEventStreamEventId since, CFTimeInterval latency, FSEventStreamCreateFlags flags) {
	context->info = (void*) info;
	return FSEventStreamCreate(NULL, (FSEventStreamCallback) fsevtCallback, context, paths, since, latency, flags);
}

static void DispatchQueueRelease(dispatch_queue_t queue) {
	dispatch_release(queue);
}
*/
import "C"

// CreateFlags specifies what events will be seen in an event stream.
type CreateFlags uint32

const (
	// NoDefer sends events on the leading edge (for interactive applications).
	// By default events are delivered after latency seconds (for background tasks).
	//
	// Affects the meaning of the EventStream.Latency parameter. If you specify
	// this flag and more than latency seconds have elapsed since
	// the last event, your app will receive the event immediately.
	// The delivery of the event resets the latency timer and any
	// further events will be delivered after latency seconds have
	// elapsed. This flag is useful for apps that are interactive
	// and want to react immediately to changes but avoid getting
	// swamped by notifications when changes are occurring in rapid
	// succession. If you do not specify this flag, then when an
	// event occurs after a period of no events, the latency timer
	// is started. Any events that occur during the next latency
	// seconds will be delivered as one group (including that first
	// event). The delivery of the group of events resets the
	// latency timer and any further events will be delivered after
	// latency seconds. This is the default behavior and is more
	// appropriate for background, daemon or batch processing apps.
	NoDefer = CreateFlags(C.kFSEventStreamCreateFlagNoDefer)

	// WatchRoot requests notifications of changes along the path to
	// the path(s) you're watching. For example, with this flag, if
	// you watch "/foo/bar" and it is renamed to "/foo/bar.old", you
	// would receive a RootChanged event. The same is true if the
	// directory "/foo" were renamed. The event you receive is a
	// special event: the path for the event is the original path
	// you specified, the flag RootChanged is set and event ID is
	// zero. RootChanged events are useful to indicate that you
	// should rescan a particular hierarchy because it changed
	// completely (as opposed to the things inside of it changing).
	// If you want to track the current location of a directory, it
	// is best to open the directory before creating the stream so
	// that you have a file descriptor for it and can issue an
	// F_GETPATH fcntl() to find the current path.
	WatchRoot = CreateFlags(C.kFSEventStreamCreateFlagWatchRoot)

	// IgnoreSelf doesn't send events triggered by the current process (macOS 10.6+).
	//
	// Don't send events that were triggered by the current process.
	// This is useful for reducing the volume of events that are
	// sent. It is only useful if your process might modify the file
	// system hierarchy beneath the path(s) being monitored. Note:
	// this has no effect on historical events, i.e., those
	// delivered before the HistoryDone sentinel event.
	IgnoreSelf = CreateFlags(C.kFSEventStreamCreateFlagIgnoreSelf)

	// FileEvents sends events about individual files, generating significantly
	// more events (macOS 10.7+) than directory level notifications.
	FileEvents = CreateFlags(C.kFSEventStreamCreateFlagFileEvents)
)

// EventFlags passed to the FSEventStreamCallback function.
// These correspond directly to the flags as described here:
// https://developer.apple.com/documentation/coreservices/1455361-fseventstreameventflags
type EventFlags uint32

const (
	// MustScanSubDirs indicates that events were coalesced hierarchically.
	//
	// Your application must rescan not just the directory given in
	// the event, but all its children, recursively. This can happen
	// if there was a problem whereby events were coalesced
	// hierarchically. For example, an event in /Users/jsmith/Music
	// and an event in /Users/jsmith/Pictures might be coalesced
	// into an event with this flag set and path=/Users/jsmith. If
	// this flag is set you may be able to get an idea of whether
	// the bottleneck happened in the kernel (less likely) or in
	// your client (more likely) by checking for the presence of the
	// informational flags UserDropped or KernelDropped.
	MustScanSubDirs EventFlags = EventFlags(C.kFSEventStreamEventFlagMustScanSubDirs)

	// KernelDropped or UserDropped may be set in addition
	// to the MustScanSubDirs flag to indicate that a problem
	// occurred in buffering the events (the particular flag set
	// indicates where the problem occurred) and that the client
	// must do a full scan of any directories (and their
	// subdirectories, recursively) being monitored by this stream.
	// If you asked to monitor multiple paths with this stream then
	// you will be notified about all of them. Your code need only
	// check for the MustScanSubDirs flag; these flags (if present)
	// only provide information to help you diagnose the problem.
	KernelDropped = EventFlags(C.kFSEventStreamEventFlagKernelDropped)

	// UserDropped is related to UserDropped above.
	UserDropped = EventFlags(C.kFSEventStreamEventFlagUserDropped)

	// EventIDsWrapped indicates the 64-bit event ID counter wrapped around.
	//
	// If EventIdsWrapped is set, it means
	// the 64-bit event ID counter wrapped around. As a result,
	// previously-issued event ID's are no longer valid
	// for the EventID field when using EventStream.Resume.
	EventIDsWrapped = EventFlags(C.kFSEventStreamEventFlagEventIdsWrapped)

	// HistoryDone is a sentinel event when retrieving events with EventStream.Resume.
	//
	// Denotes a sentinel event sent to mark the end of the
	// "historical" events sent as a result of specifying
	// EventStream.Resume.
	//
	// After sending all the "historical" events that occurred before now,
	// an event will be sent with the HistoryDone flag set. The client
	// should ignore the path supplied in that event.
	HistoryDone = EventFlags(C.kFSEventStreamEventFlagHistoryDone)

	// RootChanged indicates a change to a directory along the path being watched.
	//
	// Denotes a special event sent when there is a change to one of
	// the directories along the path to one of the directories you
	// asked to watch. When this flag is set, the event ID is zero
	// and the path corresponds to one of the paths you asked to
	// watch (specifically, the one that changed). The path may no
	// longer exist because it or one of its parents was deleted or
	// renamed. Events with this flag set will only be sent if you
	// passed the flag WatchRoot when you created the stream.
	RootChanged = EventFlags(C.kFSEventStreamEventFlagRootChanged)

	// Mount for a volume mounted underneath the path being monitored.
	//
	// Denotes a special event sent when a volume is mounted
	// underneath one of the paths being monitored. The path in the
	// event is the path to the newly-mounted volume. You will
	// receive one of these notifications for every volume mount
	// event inside the kernel (independent of DiskArbitration).
	// Beware that a newly-mounted volume could contain an
	// arbitrarily large directory hierarchy. Avoid pitfalls like
	// triggering a recursive scan of a non-local filesystem, which
	// you can detect by checking for the absence of the MNT_LOCAL
	// flag in the f_flags returned by statfs(). Also be aware of
	// the MNT_DONTBROWSE flag that is set for volumes which should
	// not be displayed by user interface elements.
	Mount = EventFlags(C.kFSEventStreamEventFlagMount)

	// Unmount event occurs after a volume is unmounted.
	//
	// Denotes a special event sent when a volume is unmounted
	// underneath one of the paths being monitored. The path in the
	// event is the path to the directory from which the volume was
	// unmounted. You will receive one of these notifications for
	// every volume unmount event inside the kernel. This is not a
	// substitute for the notifications provided by the
	// DiskArbitration framework; you only get notified after the
	// unmount has occurred. Beware that unmounting a volume could
	// uncover an arbitrarily large directory hierarchy, although
	// macOS never does that.
	Unmount = EventFlags(C.kFSEventStreamEventFlagUnmount)

	// The following flags are only set when using FileEvents.

	// ItemCreated indicates that a file or directory has been created.
	ItemCreated = EventFlags(C.kFSEventStreamEventFlagItemCreated)

	// ItemRemoved indicates that a file or directory has been removed.
	ItemRemoved = EventFlags(C.kFSEventStreamEventFlagItemRemoved)

	// ItemInodeMetaMod indicates that a file or directory's metadata has has been modified.
	ItemInodeMetaMod = EventFlags(C.kFSEventStreamEventFlagItemInodeMetaMod)

	// ItemRenamed indicates that a file or directory has been renamed.
	// TODO is there any indication what it might have been renamed to?
	ItemRenamed = EventFlags(C.kFSEventStreamEventFlagItemRenamed)

	// ItemModified indicates that a file has been modified.
	ItemModified = EventFlags(C.kFSEventStreamEventFlagItemModified)

	// ItemFinderInfoMod indicates the the item's Finder information has been
	// modified.
	// TODO the above is just a guess.
	ItemFinderInfoMod = EventFlags(C.kFSEventStreamEventFlagItemFinderInfoMod)

	// ItemChangeOwner indicates that the file has changed ownership.
	ItemChangeOwner = EventFlags(C.kFSEventStreamEventFlagItemChangeOwner)

	// ItemXattrMod indicates that the files extended attributes have changed.
	ItemXattrMod = EventFlags(C.kFSEventStreamEventFlagItemXattrMod)

	// ItemIsFile indicates that the item is a file.
	ItemIsFile = EventFlags(C.kFSEventStreamEventFlagItemIsFile)

	// ItemIsDir indicates that the item is a directory.
	ItemIsDir = EventFlags(C.kFSEventStreamEventFlagItemIsDir)

	// ItemIsSymlink indicates that the item is a symbolic link.
	ItemIsSymlink = EventFlags(C.kFSEventStreamEventFlagItemIsSymlink)
)

const (
	nullCFStringRef = C.CFStringRef(0)
	nullCFUUIDRef   = C.CFUUIDRef(0)

	// eventIDSinceNow is a sentinel to begin watching events "since now".
	eventIDSinceNow = uint64(C.kFSEventStreamEventIdSinceNow)
)

// type FSEventsCopyUUIDForDevice C.FSEventsCopyUUIDForDevice

// type CFUUIDCreateString C.CFUUIDCreateString

var kCFAllocatorDefault = C.kCFAllocatorDefault

// type FSEventsGetCurrentEventId C.FSEventsGetCurrentEventId

type FSEventStreamEventFlags C.FSEventStreamEventFlags

type FSEventStreamEventId C.FSEventStreamEventId

type fsDispatchQueueRef C.dispatch_queue_t

// fsEventStreamRef wraps C.FSEventStreamRef
type FSEventStreamRef C.FSEventStreamRef

// cfRunLoopRef wraps C.CFRunLoopRef
type cfRunLoopRef C.CFRunLoopRef

type CFStringRef C.CFStringRef

// type FSEventStreamGetLatestEventId C.FSEventStreamGetLatestEventId

// type FSEventStreamGetDeviceBeingWatched C.FSEventStreamGetDeviceBeingWatched

// type FSEventStreamCopyDescription C.FSEventStreamCopyDescription

// type FSEventStreamCopyPathsBeingWatched C.FSEventStreamCopyPathsBeingWatched

// type CFArrayGetValueAtIndex C.CFArrayGetValueAtIndex

type CFIndex C.CFIndex

// type CFStringGetLength C.CFStringGetLength

type CFStringEncoding C.CFStringEncoding

const kCFStringEncodingUTF8 = C.kCFStringEncodingUTF8

// type CFStringGetBytes C.CFStringGetBytes

type CFAbsoluteTime C.CFAbsoluteTime

// type FSEventsGetLastEventIdForDeviceBeforeTime C.FSEventsGetLastEventIdForDeviceBeforeTime

type dev C.dev_t

type CFArrayRef C.CFArrayRef

// type ArrayCreateMutable C.ArrayCreateMutable

// type CFArrayAppendValue C.CFArrayAppendValue

type CFMutableArrayRef C.CFMutableArrayRef

// type CFStringCreateWithCString C.CFStringCreateWithCString

// type CFArrayGetCount C.CFArrayGetCount

// type CFRelease C.CFRelease

type CFTypeRef C.CFTypeRef

type FSEventStreamContext C.FSEventStreamContext

type CFTimeInterval C.CFTimeInterval

// type EventStreamCreateRelativeToDevice C.EventStreamCreateRelativeToDevice

type FSEventStreamCreateFlags C.FSEventStreamCreateFlags

// type EventStreamCreate C.EventStreamCreate

// type dispatch_queue_create C.dispatch_queue_create

// type FSEventStreamSetDispatchQueue C.FSEventStreamSetDispatchQueue

// type FSEventStreamStart C.FSEventStreamStart

// type FSEventStreamInvalidate C.FSEventStreamInvalidate

// type FSEventStreamRelease C.FSEventStreamRelease

// type DispatchQueueRelease C.DispatchQueueRelease

// type FSEventStreamFlushSync C.FSEventStreamFlushSync

// type FSEventStreamFlushAsync C.FSEventStreamFlushAsync

// type FSEventStreamStop C.FSEventStreamStop
