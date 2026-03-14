'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { useUpdateProfile, useChangePassword, useDeleteAccount } from '@/hooks/use-lists';

export default function SettingsPage() {
  const { user } = useAuthStore();
  const router = useRouter();
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();
  const deleteAccount = useDeleteAccount();

  // Profile form
  const [name, setName] = useState(user?.name ?? '');
  const [currentTitle, setCurrentTitle] = useState(user?.current_title ?? '');
  const [experienceLevel, setExperienceLevel] = useState(user?.experience_level ?? '');
  const [profileSaved, setProfileSaved] = useState(false);

  // Password form
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [passwordChanged, setPasswordChanged] = useState(false);

  // Delete account
  const [deletePassword, setDeletePassword] = useState('');
  const [showDelete, setShowDelete] = useState(false);

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await updateProfile.mutateAsync({
      name: name.trim() || undefined,
      current_title: currentTitle.trim() || undefined,
      experience_level: experienceLevel || undefined,
    });
    setProfileSaved(true);
    setTimeout(() => setProfileSaved(false), 3000);
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await changePassword.mutateAsync({
      current_password: currentPassword,
      new_password: newPassword,
    });
    setCurrentPassword('');
    setNewPassword('');
    setPasswordChanged(true);
    setTimeout(() => setPasswordChanged(false), 3000);
  };

  const handleDeleteAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!confirm('Are you sure? Your account will be permanently deleted after 30 days.')) return;
    await deleteAccount.mutateAsync({ password: deletePassword });
    router.push('/login');
  };

  return (
    <div className="space-y-10">
      <div>
        <h1 className="text-2xl font-bold text-slate-100">Settings</h1>
        <p className="mt-1 text-sm text-slate-500">Manage your profile and account</p>
      </div>

      {/* Profile section */}
      <section className="rounded-lg border border-edge bg-card p-6">
        <h2 className="text-lg font-semibold text-slate-100">Profile</h2>
        <form onSubmit={handleProfileSubmit} className="mt-4 space-y-4">
          <div>
            <label htmlFor="settings-name" className="block text-sm font-medium text-slate-300">
              Name
            </label>
            <input
              id="settings-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
          </div>
          <div>
            <label htmlFor="settings-email" className="block text-sm font-medium text-slate-300">
              Email
            </label>
            <input
              id="settings-email"
              type="email"
              value={user?.email ?? ''}
              disabled
              className="mt-1 block w-full max-w-md rounded-md border border-edge bg-overlay px-3 py-2 text-sm text-slate-500"
            />
          </div>
          <div>
            <label htmlFor="settings-title" className="block text-sm font-medium text-slate-300">
              Current Title
            </label>
            <input
              id="settings-title"
              type="text"
              value={currentTitle}
              onChange={(e) => setCurrentTitle(e.target.value)}
              placeholder="e.g., Senior Software Engineer"
              className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
          </div>
          <div>
            <label htmlFor="settings-level" className="block text-sm font-medium text-slate-300">
              Experience Level
            </label>
            <select
              id="settings-level"
              value={experienceLevel}
              onChange={(e) => setExperienceLevel(e.target.value)}
              className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            >
              <option value="">Select...</option>
              <option value="fresher">Fresher</option>
              <option value="junior">Junior</option>
              <option value="mid">Mid-level</option>
              <option value="senior">Senior</option>
              <option value="staff_plus">Staff+</option>
            </select>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={updateProfile.isPending}
              className="btn-neon rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {updateProfile.isPending ? 'Saving...' : 'Save profile'}
            </button>
            {profileSaved && (
              <span className="text-sm text-green-400">Saved!</span>
            )}
            {updateProfile.isError && (
              <span className="text-sm text-red-400">
                {(updateProfile.error as Error).message}
              </span>
            )}
          </div>
        </form>
      </section>

      {/* Password section */}
      <section className="rounded-lg border border-edge bg-card p-6">
        <h2 className="text-lg font-semibold text-slate-100">Change Password</h2>
        <form onSubmit={handlePasswordSubmit} className="mt-4 space-y-4">
          <div>
            <label htmlFor="current-pw" className="block text-sm font-medium text-slate-300">
              Current Password
            </label>
            <input
              id="current-pw"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
          </div>
          <div>
            <label htmlFor="new-pw" className="block text-sm font-medium text-slate-300">
              New Password
            </label>
            <input
              id="new-pw"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
            <p className="mt-1 text-xs text-slate-600">
              Min 8 characters, 1 uppercase, 1 lowercase, 1 digit
            </p>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={!currentPassword || !newPassword || changePassword.isPending}
              className="btn-neon rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {changePassword.isPending ? 'Changing...' : 'Change password'}
            </button>
            {passwordChanged && (
              <span className="text-sm text-green-400">Password changed!</span>
            )}
            {changePassword.isError && (
              <span className="text-sm text-red-400">
                {(changePassword.error as Error).message}
              </span>
            )}
          </div>
        </form>
      </section>

      {/* Danger zone */}
      <section className="rounded-lg border border-red-900/50 bg-card p-6">
        <h2 className="text-lg font-semibold text-red-400">Danger Zone</h2>
        <p className="mt-1 text-sm text-slate-500">
          Deleting your account is irreversible. Your data will be permanently removed after 30 days.
        </p>
        {!showDelete ? (
          <button
            onClick={() => setShowDelete(true)}
            className="mt-4 rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-400 hover:bg-red-50"
          >
            Delete my account
          </button>
        ) : (
          <form onSubmit={handleDeleteAccount} className="mt-4 space-y-3">
            <div>
              <label htmlFor="delete-pw" className="block text-sm font-medium text-slate-300">
                Confirm your password
              </label>
              <input
                id="delete-pw"
                type="password"
                value={deletePassword}
                onChange={(e) => setDeletePassword(e.target.value)}
                className="mt-1 block w-full max-w-md rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
              />
            </div>
            <div className="flex gap-3">
              <button
                type="submit"
                disabled={!deletePassword || deleteAccount.isPending}
                className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                {deleteAccount.isPending ? 'Deleting...' : 'Permanently delete'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowDelete(false);
                  setDeletePassword('');
                }}
                className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-slate-300 hover:bg-overlay"
              >
                Cancel
              </button>
            </div>
            {deleteAccount.isError && (
              <p className="text-sm text-red-400">
                {(deleteAccount.error as Error).message}
              </p>
            )}
          </form>
        )}
      </section>
    </div>
  );
}
