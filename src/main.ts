import { NestFactory } from "@nestjs/core";
import { AppModule } from "./app/app.module";
import {
  FastifyAdapter,
  NestFastifyApplication,
} from "@nestjs/platform-fastify";
import { SwaggerModule, DocumentBuilder } from "@nestjs/swagger";
import { ServeStaticModule } from "@nestjs/serve-static";

async function bootstrap() {
  const app = await NestFactory.create<NestFastifyApplication>(
    AppModule,
    new FastifyAdapter({ logger: true })
  );
  const config = new DocumentBuilder()
    .setTitle("Tiny ILS")
    .setDescription("The Tiny ILS API has many functions which you can use.")
    .setVersion("1.0")
    .build();
  const documentFactory = () => SwaggerModule.createDocument(app, config);
  SwaggerModule.setup("api", app, documentFactory, {
    jsonDocumentUrl: "swagger/json",
  });
  await app.listen(process.env.PORT, "0.0.0.0");
}
bootstrap();
