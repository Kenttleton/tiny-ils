import { Module } from "@nestjs/common";
import { ApiController } from "../api/api.controller";
import { ApiService } from "../api/api.service";
import { ConfigModule } from "@nestjs/config";
import { TypeOrmModule } from "@nestjs/typeorm";
import { HttpModule } from "@nestjs/axios";
import { AuthModule } from "../auth/auth.module";
import { UsersService } from "../users/users.service";
import { UsersModule } from "../users/users.module";
import { AuthController } from "src/auth/auth.controller";
import { ServeStaticModule } from "@nestjs/serve-static";
import { join } from "path";

@Module({
  imports: [
    ConfigModule.forRoot(),
    HttpModule,
    AuthModule,
    UsersModule,
    // TypeOrmModule.forRoot({
    //   type: "mysql",
    //   host: "localhost",
    //   port: 3306,
    //   username: "root",
    //   password: "root",
    //   database: "test",
    //   entities: [],
    //   synchronize: true,
    // }),
    ServeStaticModule.forRoot({
      rootPath: join(__dirname, "../..", "frontend/dist"),
      exclude: ["/api/(.*)", "/auth/(.*)"],
    }),
  ],
  controllers: [ApiController, AuthController],
  providers: [ApiService, UsersService],
})
export class AppModule {}
