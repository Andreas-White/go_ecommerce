function validateEmail(email: string): string | null {
    const trimmedEmail = email.trim();

    if (!trimmedEmail) {
      return 'Email is required';
    }

    if (trimmedEmail.length > 128) {
      return 'Email address is too long';
    }

    const emailRegex =
      /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

    const [localPart, domain] = trimmedEmail.split('@');
    if (localPart.length > 64) {
      return 'Email address is invalid';
    }

    if (trimmedEmail.includes('..')) {
      return 'Email address cannot contain consecutive dots';
    }

    if (localPart.startsWith('.') || localPart.endsWith('.')) {
      return 'Email address cannot start or end with a dot';
    }

    if (!emailRegex.test(trimmedEmail)) {
      return 'Please enter a valid email address';
    }

    return null;
}

export default validateEmail;