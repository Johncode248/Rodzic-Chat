package chatspg

type ChannelTransient struct {
	Previous Channel
	Current  Channel
	Future   Channel
}

func (t ChannelTransient) Update() ChannelTransient {
	return ChannelTransient{Previous: t.Current, Current: t.Future, Future: Channel{}}
}

type MessageTransient struct {
	Previous Message
	Current  Message
	Future   Message
}

func (t MessageTransient) Update() MessageTransient {
	return MessageTransient{Previous: t.Current, Current: t.Future, Future: Message{}}
}

type PollTransient struct {
	Previous Poll
	Current  Poll
	Future   Poll
}

func (t PollTransient) Update() PollTransient {
	return PollTransient{Previous: t.Current, Current: t.Future, Future: Poll{}}
}

type PollOptionTransient struct {
	Previous PollOption
	Current  PollOption
	Future   PollOption
}

func (t PollOptionTransient) Update() PollOptionTransient {
	return PollOptionTransient{Previous: t.Current, Current: t.Future, Future: PollOption{}}
}

type PollVoteTransient struct {
	Previous PollVote
	Current  PollVote
	Future   PollVote
}

func (t PollVoteTransient) Update() PollVoteTransient {
	return PollVoteTransient{Previous: t.Current, Current: t.Future, Future: PollVote{}}
}

type MessageReadTransient struct {
	Previous MessageRead
	Current  MessageRead
	Future   MessageRead
}

func (t MessageReadTransient) Update() MessageReadTransient {
	return MessageReadTransient{Previous: t.Current, Current: t.Future, Future: MessageRead{}}
}

type MessageReactionTransient struct {
	Previous MessageReaction
	Current  MessageReaction
	Future   MessageReaction
}

func (t MessageReactionTransient) Update() MessageReactionTransient {
	return MessageReactionTransient{Previous: t.Current, Current: t.Future, Future: MessageReaction{}}
}
