import { Controller, Get, Post, UseGuards, Request } from "@nestjs/common";
import { LocalAuthGuard } from "src/auth/local-auth.guard";
import { AuthService } from "src/auth/auth.service";

@Controller("auth")
export class AuthController {
  constructor(private authService: AuthService) {}

  @UseGuards(LocalAuthGuard)
  @Post("login")
  async login(@Request() req) {
    return this.authService.login(req.user);
  }

  @UseGuards(LocalAuthGuard)
  @Post("logout")
  async logout(@Request() req) {
    return req.logout();
  }
}
