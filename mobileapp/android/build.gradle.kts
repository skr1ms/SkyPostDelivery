allprojects {
    repositories {
        google()
        mavenCentral()
        
        // Зеркала для стабильности
        maven { url = uri("https://maven.aliyun.com/repository/public") }
        maven { url = uri("https://maven.aliyun.com/repository/google") }
        maven { url = uri("https://jitpack.io") }
    }
}

val newBuildDir: Directory =
    rootProject.layout.buildDirectory
        .dir("../../build")
        .get()
rootProject.layout.buildDirectory.value(newBuildDir)

subprojects {
    val newSubprojectBuildDir: Directory = newBuildDir.dir(project.name)
    project.layout.buildDirectory.value(newSubprojectBuildDir)
}
/*subprojects {
    project.evaluationDependsOn(":app")
    
    afterEvaluate {
        if (project.hasProperty("android")) {
            val android = project.extensions.findByName("android")
            if (android != null) {
                try {
                    val compileOptions = (android as Any).javaClass.getMethod("getCompileOptions").invoke(android)
                    compileOptions?.javaClass?.getMethod("setSourceCompatibility", Any::class.java)?.invoke(compileOptions, org.gradle.api.JavaVersion.VERSION_11)
                    compileOptions?.javaClass?.getMethod("setTargetCompatibility", Any::class.java)?.invoke(compileOptions, org.gradle.api.JavaVersion.VERSION_11)
                } catch (e: Exception) {
                }
            }
        }
    }
}
*/

tasks.register<Delete>("clean") {
    delete(rootProject.layout.buildDirectory)
}
