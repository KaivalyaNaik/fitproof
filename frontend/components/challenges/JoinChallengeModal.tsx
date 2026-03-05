"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import { Modal } from "@/components/ui/Modal";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";

interface JoinChallengeModalProps {
  open: boolean;
  onClose: () => void;
}

export function JoinChallengeModal({ open, onClose }: JoinChallengeModalProps) {
  const router = useRouter();
  const [inviteCode, setInviteCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const challenge = await clientFetch<{ id: string }>("/challenges/join", {
        method: "POST",
        body: JSON.stringify({ invite_code: inviteCode.trim().toUpperCase() }),
      });
      setInviteCode("");
      onClose();
      router.push(`/challenges/${challenge.id}`);
    } catch (err) {
      if (err instanceof ApiResponseError) {
        if (err.status === 409) setError("You are already a member of this challenge.");
        else if (err.status === 404) setError("Invalid invite code.");
        else setError(err.message);
      } else {
        setError("Something went wrong.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Join a Challenge">
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input
          id="invite_code"
          label="Invite code"
          placeholder="e.g. PYMQHZX1"
          value={inviteCode}
          onChange={(e) => setInviteCode(e.target.value)}
          required
          className="font-mono uppercase"
        />

        {error && (
          <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2">
            {error}
          </p>
        )}

        <div className="flex justify-end gap-2 mt-1">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Join
          </Button>
        </div>
      </form>
    </Modal>
  );
}
