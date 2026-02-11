import type { MigrationInterface, QueryRunner } from 'typeorm';

export class AddVerticalsAndFormats1708100000000 implements MigrationInterface {
  name = 'AddVerticalsAndFormats1708100000000';

  public async up(queryRunner: QueryRunner): Promise<void> {
    // 1. Campaigns - Add Vertical
    await queryRunner.query(
      `ALTER TABLE "campaigns" ADD "vertical" character varying`,
    );

    // 2. Ad Units - Update Enum (Postgres workaround for adding values)
    // We change column to text temporarily, drop type, recreate type, cast back
    await queryRunner.query(
      `ALTER TABLE "ad_units" ALTER COLUMN "type" TYPE text`,
    );
    await queryRunner.query(`DROP TYPE "public"."ad_units_type_enum"`);
    await queryRunner.query(
      `CREATE TYPE "public"."ad_units_type_enum" AS ENUM(
        'banner', 'rich_media', 
        'video_instream', 'video_outstream', 'ctv', 
        'native', 'content_recommendation', 
        'interstitial', 'rewarded', 'playable', 
        'audio_digital', 'audio_programmatic', 
        'dco', 'vr_ar', 'in_game', 
        'push', 'popunder'
      )`,
    );
    await queryRunner.query(
      `ALTER TABLE "ad_units" ALTER COLUMN "type" TYPE "public"."ad_units_type_enum" USING "type"::"public"."ad_units_type_enum"`,
    );

    // 3. Audience Segments - Update Enum
    await queryRunner.query(
      `ALTER TABLE "dsp_audience_segments" ALTER COLUMN "type" TYPE text`,
    );
    await queryRunner.query(`DROP TYPE "public"."dsp_audience_segments_type_enum"`);
    await queryRunner.query(
      `CREATE TYPE "public"."dsp_audience_segments_type_enum" AS ENUM(
        'first_party', 'third_party', 'lookalike', 'contextual', 'retargeting',
        'demographic', 'psychographic', 'behavioral', 'b2b', 'intent'
      )`,
    );
    await queryRunner.query(
      `ALTER TABLE "dsp_audience_segments" ALTER COLUMN "type" TYPE "public"."dsp_audience_segments_type_enum" USING "type"::"public"."dsp_audience_segments_type_enum"`,
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Revert Vertical
    await queryRunner.query(`ALTER TABLE "campaigns" DROP COLUMN "vertical"`);

    // Revert Ad Units (Simplified revert to basic types)
    await queryRunner.query(
      `ALTER TABLE "ad_units" ALTER COLUMN "type" TYPE text`,
    );
    await queryRunner.query(`DROP TYPE "public"."ad_units_type_enum"`);
    await queryRunner.query(
      `CREATE TYPE "public"."ad_units_type_enum" AS ENUM('banner', 'video', 'native', 'interstitial', 'rewarded')`,
    );
    await queryRunner.query(
      `ALTER TABLE "ad_units" ALTER COLUMN "type" TYPE "public"."ad_units_type_enum" USING "type"::"public"."ad_units_type_enum"`,
    );

    // Revert Audience Segments
    await queryRunner.query(
      `ALTER TABLE "dsp_audience_segments" ALTER COLUMN "type" TYPE text`,
    );
    await queryRunner.query(`DROP TYPE "public"."dsp_audience_segments_type_enum"`);
    await queryRunner.query(
      `CREATE TYPE "public"."dsp_audience_segments_type_enum" AS ENUM('first_party', 'third_party', 'lookalike', 'contextual', 'retargeting')`,
    );
    await queryRunner.query(
      `ALTER TABLE "dsp_audience_segments" ALTER COLUMN "type" TYPE "public"."dsp_audience_segments_type_enum" USING "type"::"public"."dsp_audience_segments_type_enum"`,
    );
  }
}
