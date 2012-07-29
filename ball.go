package main

import "mud"

type Ball struct { mud.PhysicalObject }

func (b Ball) Visible() bool { return true }
func (b Ball) Description() string { return "A &red;red&; ball" }
func (b Ball) Carryable() bool { return true }
func (b Ball) TextHandles() []string { return []string{"ball","red ball"} }