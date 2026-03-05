"use client";

import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { JoinChallengeModal } from "./JoinChallengeModal";

export function JoinChallengeButton() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button variant="secondary" size="sm" onClick={() => setOpen(true)}>
        Join Challenge
      </Button>
      <JoinChallengeModal open={open} onClose={() => setOpen(false)} />
    </>
  );
}
