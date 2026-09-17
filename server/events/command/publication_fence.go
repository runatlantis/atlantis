// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

// PublicationWriteMode makes every durable writer explicitly choose a fence or
// intentional unfenced access. Backends reject NoClaim while a lease is held.
type PublicationWriteMode interface{ publicationWriteMode() }

type PublicationFence struct{ Owner string }

func (PublicationFence) publicationWriteMode() {}

type NoClaim struct{}

func (NoClaim) publicationWriteMode() {}
