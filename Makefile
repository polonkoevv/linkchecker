# ==============================================================================
# Makefile for Project "Link checker"
#
# Description:
#   This Makefile handles the build process for the link checker application,
#   including compilation, testing, packaging, and cleanup.
#
# Usage:
#
# Author: Bersnakx
# ==============================================================================

SRCDIR :=./cmd/main.go


.PHONY: run

run:
	go run $(SRCDIR)
