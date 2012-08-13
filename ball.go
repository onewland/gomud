package main

import "mud"

type Ball struct { mud.PhysicalObject; r *mud.Room }

func (b Ball) Visible() bool { return true }
func (b Ball) Description() string { return "A &red;red&; ball" }
func (b Ball) Carryable() bool { return true }
func (b Ball) TextHandles() []string { return []string{"ball","red ball"} }
func (b *Ball) SetRoom(r *mud.Room) { b.r = r }
func (b Ball) Room() *mud.Room { return b.r }