import { SetMetadata } from '@nestjs/common';
import type { UserRole } from '../../users/user.entity';

export const Roles = (...roles: UserRole[]) => SetMetadata('roles', roles);
