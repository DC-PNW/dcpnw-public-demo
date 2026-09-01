plugins {
    kotlin("jvm") version "1.3.50"
}

repositories {
    mavenCentral()
}

dependencies {
    implementation("org.springframework:spring-core:4.3.0.RELEASE")
    implementation("org.apache.logging.log4j:log4j-core:2.14.1")
    implementation("commons-collections:commons-collections:3.2.1")
    implementation("org.apache.commons:commons-configuration2:2.4")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.9.0")
    implementation("mysql:mysql-connector-java:5.1.49")
    implementation("org.apache.struts:struts2-core:2.3.1")
    implementation("ch.qos.logback:logback-core:1.2.3")
    implementation("org.jboss.marshalling:jboss-marshalling:1.4.10.Final")
    implementation("org.bouncycastle:bcprov-jdk15on:1.64")
    implementation("com.google.guava:guava:20.0")
    implementation("org.jsoup:jsoup:1.10.2")
    implementation("org.apache.pdfbox:pdfbox:2.0.4")
    implementation("org.codehaus.groovy:groovy:2.4.5")

    testImplementation("junit:junit:4.12")
    testImplementation("org.mockito:mockito-core:2.28.2")
    testImplementation("org.seleniumhq.selenium:selenium-java:3.141.59")
    testImplementation("org.apache.httpcomponents:httpclient:4.3.6")
    testImplementation("io.rest-assured:rest-assured:3.0.0")
    testImplementation("com.h2database:h2:1.4.197")
    testImplementation("org.apache.maven:maven-artifact:3.1.1")
    testImplementation("org.apache.poi:poi:3.17")
    testImplementation("org.jetbrains.kotlin:kotlin-scripting-compiler-embeddable:1.3.50")
}

